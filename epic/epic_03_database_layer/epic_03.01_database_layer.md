# Epic 03 — Database Layer

**Goal:** Establish PostgreSQL connection via pgx/v5, set up GORM with proper connection pooling, implement golang-migrate integration.
**Dependencies:** Epic 02 (Config loaded)
**Commit:** `feat: add PostgreSQL database layer with GORM and migrations`

---

## Step 03.01 — PostgreSQL Connection with pgx/v5

**Build:** Create `backend/internal/infrastructure/database/postgres.go`:

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/muhammadyunus/ForgeBase/internal/config"
)

// PostgresDB wraps pgx pool with the repository.DB interface.
type PostgresDB struct {
    pool *pgxpool.Pool
}

// NewPostgresDB creates a new PostgreSQL connection pool.
func NewPostgresDB(ctx context.Context, cfg config.DatabaseConfig) (*PostgresDB, error) {
    poolConfig, err := pgxpool.ParseConfig(cfg.URL)
    if err != nil {
        return nil, fmt.Errorf("parse database config: %w", err)
    }

    // Connection pool settings
    poolConfig.MaxConns = 25
    poolConfig.MinConns = 5
    poolConfig.MaxConnLifetime = 5 * time.Minute
    poolConfig.MaxConnIdleTime = 1 * time.Minute
    poolConfig.HealthCheckPeriod = 30 * time.Second

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    // Verify connection
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return &PostgresDB{pool: pool}, nil
}

// Pool returns the underlying pgxpool for GORM integration.
func (p *PostgresDB) Pool() *pgxpool.Pool {
    return p.pool
}

// SQLDB returns a standard *sql.DB wrapper for compatibility.
func (p *PostgresDB) SQLDB() *sql.DB {
    return p.pool.DB()
}

// Close shuts down the connection pool.
func (p *PostgresDB) Close(_ context.Context) error {
    p.pool.Close()
    return nil
}

// DB interface implementation
var _ DB = (*PostgresDB)(nil)

// DB is the database repository interface.
type DB interface {
    Pool() *pgxpool.Pool
    SQLDB() *sql.DB
    Close(ctx context.Context) error
    Ping(ctx context.Context) error
}

func (p *PostgresDB) Ping(ctx context.Context) error {
    return p.pool.Ping(ctx)
}

// gormDB extracts the GORM instance from a RepositoryDB wrapper.
// NOTE: GORM instance is stored in the DI container and retrieved via
// a package-level variable set during bootstrap.
var gormDBInstance *gorm.DB

// SetGORMDB sets the global GORM instance for repository access.
func SetGORMDB(db *gorm.DB) {
    gormDBInstance = db
}

// GORM returns the global GORM instance. Must be called after bootstrap.
func GORM() *gorm.DB {
    return gormDBInstance
}
```

**Test cases:**
- [ ] Unit: `NewPostgresDB()` with valid DSN creates pool successfully
- [ ] Unit: `NewPostgresDB()` with invalid DSN returns error
- [ ] Unit: `Close()` closes pool without error
- [ ] Integration: Connection pool health check passes

---

## Step 03.02 — GORM Integration

**Build:** Create `backend/internal/infrastructure/database/gorm.go`:

```go
package database

import (
    "fmt"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// GORMDB wraps GORM with the underlying pgx connection.
type GORMDB struct {
    db *gorm.DB
}

// NewGORMDB creates a GORM instance connected to PostgreSQL.
func NewGORMDB(pg *PostgresDB) (*GORMDB, error) {
    db, err := gorm.Open(postgres.Open(pg.Pool().Config.ConnString()), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("open gorm: %w", err)
    }

    sqlDB := pg.SQLDB()
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    return &GORMDB{db: db}, nil
}

// DB returns the raw GORM *gorm.DB for repository implementations.
func (g *GORMDB) DB() *gorm.DB {
    return g.db
}

// Close delegates to the underlying postgres connection.
func (g *GORMDB) Close(ctx context.Context) error {
    sqlDB, err := g.db.DB()
    if err != nil {
        return err
    }
    return sqlDB.Close()
}
```

---

## Step 03.03 — Migration System

**Build:** Create `backend/migrations/` directory structure:

```
migrations/
├── 000001_create_users_table.down.sql
├── 000001_create_users_table.up.sql
├── 000002_create_roles_table.down.sql
├── 000002_create_roles_table.up.sql
├── 000003_create_user_roles_table.down.sql
├── 000003_create_user_roles_table.up.sql
├── 000004_create_workspaces_table.down.sql
├── 000004_create_workspaces_table.up.sql
├── 000005_create_teams_table.down.sql
├── 000005_create_teams_table.up.sql
├── 000006_create_team_memberships_table.down.sql
├── 000006_create_team_memberships_table.up.sql
├── 000007_create_workspace_teams_table.down.sql
├── 000007_create_workspace_teams_table.up.sql
├── 000008_create_collections_table.down.sql
├── 000008_create_collections_table.up.sql
├── 000009_create_endpoints_table.down.sql
├── 000009_create_endpoints_table.up.sql
├── 000010_create_api_logs_table.down.sql
├── 000010_create_api_logs_table.up.sql
├── 000011_create_api_analytics_table.down.sql
├── 000011_create_api_analytics_table.up.sql
├── 000012_create_alerts_table.down.sql
├── 000012_create_alerts_table.up.sql
├── 000013_create_jwt_blacklist_table.down.sql
├── 000013_create_jwt_blacklist_table.up.sql
```

**Example migration file** `000001_create_users_table.up.sql`:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Example migration file** `000001_create_users_table.down.sql`:
```sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

**Build:** Create migration runner `backend/internal/infrastructure/database/migrator.go`:

```go
package database

import (
    "context"
    "embed"
    "fmt"
    "io/fs"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed ../../migrations/*.sql
var migrationFS embed.FS

// Migrator applies database migrations.
type Migrator struct {
    m *migrate.Migrate
}

// NewMigrator creates a migrator for the given database connection.
func NewMigrator(pg *PostgresDB) (*Migrator, error) {
    source, err := iofs.New(migrationFS, "migrations")
    if err != nil {
        return nil, fmt.Errorf("load migrations: %w", err)
    }

    driver, err := postgres.WithInstance(pg.SQLDB(), &postgres.Config{})
    if err != nil {
        return nil, fmt.Errorf("create postgres driver: %w", err)
    }

    m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
    if err != nil {
        return nil, fmt.Errorf("create migrator: %w", err)
    }

    return &Migrator{m: m}, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up(ctx context.Context) error {
    if err := m.m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("migrate up: %w", err)
    }
    return nil
}

// Version returns the current migration version.
func (m *Migrator) Version(ctx context.Context) (uint, bool, error) {
    return m.m.Version()
}

// ListVersions returns all migration versions.
func (m *Migrator) ListVersions(ctx context.Context) ([]uint, error) {
    versions, err := m.m.Steps(1000)
    if err != nil {
        return nil, err
    }
    result := make([]uint, len(versions))
    for i, v := range versions {
        result[i] = v
    }
    return result, nil
}
```

**Test cases:**
- [ ] Unit: Migration files are embedded correctly
- [ ] Integration: `migrator.Up()` applies all migrations to test database
- [ ] Integration: Running migrations twice is idempotent (no error on second run)
- [ ] Integration: Down migration removes tables correctly

---

## Step 03.04 — Database Repository Interface

**Build:** Create `backend/internal/domain/repository/database.go`:

```go
package repository

import "context"

// DB is the database repository interface.
// All infrastructure adapters implement this.
type DB interface {
    // BeginTransaction starts a new database transaction.
    BeginTransaction(ctx context.Context) (Transaction, error)

    // WithTransaction executes fn within a transaction.
    WithTransaction(ctx context.Context, fn func(tx Transaction) error) error

    // Raw executes raw SQL and scans into destination.
    Raw(ctx context.Context, query string, dest any, args ...any) error

    // Query executes a query and returns rows.
    Query(ctx context.Context, query string, args ...any) ([]map[string]any, error)

    // Close closes the database connection.
    Close(ctx context.Context) error
}

// Transaction represents a database transaction.
type Transaction interface {
    DB
    Commit() error
    Rollback() error
}
```

**Test cases:**
- [ ] Unit: Interface satisfies compile-time check
- [ ] Integration: Transaction commits correctly
- [ ] Integration: Transaction rolls back on error
- [ ] Integration: Nested transactions use savepoints

---

## Step 03.05 — Update DI Bootstrap with Database

**Build:** Update `internal/di/bootstrap.go` — replace `initDatabase` stub with real call:

```go
func initDatabase(ctx context.Context, cfg config.DatabaseConfig) (repository.DB, error) {
    pg, err := database.NewPostgresDB(ctx, cfg)
    if err != nil {
        return nil, err
    }
    _, err = database.NewGORMDB(pg)
    if err != nil {
        pg.Close(ctx)
        return nil, err
    }
    return pg, nil
}
```

**Test cases:**
- [ ] Integration: Full bootstrap with real PostgreSQL connection
- [ ] E2E: Application starts with database connected

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add PostgreSQL database layer with pgx/v5, GORM, and golang-migrate"
```
