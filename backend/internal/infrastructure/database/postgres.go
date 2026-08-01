package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// PostgresDB wraps a pgx connection pool and implements repository.DB.
type PostgresDB struct {
	pool *pgxpool.Pool
}

// NewPostgresDB creates a lazily-connected PostgreSQL connection pool.
// The pool is validated with Ping by the caller (DI bootstrap) so that
// construction remains testable without a running database.
func NewPostgresDB(ctx context.Context, cfg config.DatabaseConfig) (*PostgresDB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 1 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// Pool returns the underlying pgx pool for GORM integration.
func (p *PostgresDB) Pool() *pgxpool.Pool {
	return p.pool
}

// SQLDB returns a standard *sql.DB wrapper sharing the pool.
func (p *PostgresDB) SQLDB() *sql.DB {
	return stdlib.OpenDBFromPool(p.pool)
}

// Ping verifies the database connection.
func (p *PostgresDB) Ping(ctx context.Context) error {
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}

// Close shuts down the connection pool.
func (p *PostgresDB) Close(_ context.Context) error {
	p.pool.Close()

	return nil
}

// BeginTransaction starts a new database transaction.
func (p *PostgresDB) BeginTransaction(ctx context.Context) (repository.Transaction, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	return &pgxTx{tx: tx, ctx: ctx}, nil
}

// WithTransaction executes fn within a transaction, committing on success
// and rolling back on error.
func (p *PostgresDB) WithTransaction(ctx context.Context, fn func(tx repository.Transaction) error) error {
	tx, err := p.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback transaction: %w", errors.Join(err, fmt.Errorf("rollback: %w", rbErr)))
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// Raw executes raw SQL and scans the result into dest.
func (p *PostgresDB) Raw(ctx context.Context, query string, dest any, args ...any) error {
	return scanInto(ctx, p.pool, query, dest, args...)
}

// Query executes a query and returns the rows as maps keyed by column name.
func (p *PostgresDB) Query(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	return queryRows(ctx, p.pool, query, args...)
}

// pgxTx implements repository.Transaction over a pgx.Tx, using savepoints
// for nested transactions.
type pgxTx struct {
	tx        pgx.Tx
	ctx       context.Context
	savepoint string
}

var savepointSeq atomic.Uint64

func (t *pgxTx) BeginTransaction(ctx context.Context) (repository.Transaction, error) {
	name := fmt.Sprintf("forgebase_sp_%d", savepointSeq.Add(1))

	if _, err := t.tx.Exec(ctx, "SAVEPOINT "+name); err != nil {
		return nil, fmt.Errorf("create savepoint: %w", err)
	}

	return &pgxTx{tx: t.tx, ctx: ctx, savepoint: name}, nil
}

func (t *pgxTx) WithTransaction(ctx context.Context, fn func(tx repository.Transaction) error) error {
	tx, err := t.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback savepoint: %w", errors.Join(err, fmt.Errorf("rollback: %w", rbErr)))
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit savepoint: %w", err)
	}

	return nil
}

func (t *pgxTx) Raw(ctx context.Context, query string, dest any, args ...any) error {
	return scanInto(ctx, t.tx, query, dest, args...)
}

func (t *pgxTx) Query(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	return queryRows(ctx, t.tx, query, args...)
}

//nolint:contextcheck // Transaction interface has no ctx on Close; context captured at begin.
func (t *pgxTx) Close(_ context.Context) error {
	if t.savepoint != "" {
		if _, err := t.tx.Exec(t.ctx, "ROLLBACK TO SAVEPOINT "+t.savepoint); err != nil {
			return fmt.Errorf("rollback to savepoint: %w", err)
		}

		return nil
	}

	if err := t.tx.Rollback(t.ctx); err != nil {
		return fmt.Errorf("rollback transaction: %w", err)
	}

	return nil
}

//nolint:contextcheck // Transaction interface has no ctx on Commit; context captured at begin.
func (t *pgxTx) Commit() error {
	if t.savepoint != "" {
		if _, err := t.tx.Exec(t.ctx, "RELEASE SAVEPOINT "+t.savepoint); err != nil {
			return fmt.Errorf("release savepoint: %w", err)
		}

		return nil
	}

	if err := t.tx.Commit(t.ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

//nolint:contextcheck // Transaction interface has no ctx on Rollback; context captured at begin.
func (t *pgxTx) Rollback() error {
	if t.savepoint != "" {
		if _, err := t.tx.Exec(t.ctx, "ROLLBACK TO SAVEPOINT "+t.savepoint); err != nil {
			return fmt.Errorf("rollback to savepoint: %w", err)
		}

		return nil
	}

	if err := t.tx.Rollback(t.ctx); err != nil {
		return fmt.Errorf("rollback transaction: %w", err)
	}

	return nil
}

// rowQuerier is satisfied by both pgxpool.Pool and pgx.Tx.
type rowQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func scanInto(ctx context.Context, q rowQuerier, query string, dest any, args ...any) error {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	switch d := dest.(type) {
	case *[]map[string]any:
		result, err := pgx.CollectRows(rows, pgx.RowToMap)
		if err != nil {
			return fmt.Errorf("scan rows: %w", err)
		}

		*d = result

		return nil

	case *map[string]any:
		result, err := pgx.CollectRows(rows, pgx.RowToMap)
		if err != nil {
			return fmt.Errorf("scan rows: %w", err)
		}

		if len(result) == 0 {
			return pgx.ErrNoRows
		}

		*d = result[0]

		return nil

	default:
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return fmt.Errorf("iterate rows: %w", err)
			}

			return pgx.ErrNoRows
		}

		if err := rows.Scan(dest); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}

		return nil
	}
}

func queryRows(ctx context.Context, q rowQuerier, query string, args ...any) ([]map[string]any, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	result, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return nil, fmt.Errorf("scan rows: %w", err)
	}

	return result, nil
}

var _ repository.DB = (*PostgresDB)(nil)
