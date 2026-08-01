package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/muhammadyunus/Restify-Service/migrations"
)

// Migrator applies database migrations.
type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator creates a migrator for the given database connection.
func NewMigrator(pg *PostgresDB) (*Migrator, error) {
	source, err := iofs.New(migrations.FS, ".")
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
func (m *Migrator) Up(_ context.Context) error {
	if err := m.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

// Version returns the current migration version.
func (m *Migrator) Version(_ context.Context) (uint, bool, error) {
	version, dirty, err := m.m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("get migration version: %w", err)
	}

	return version, dirty, nil
}

// ListVersions returns all available migration versions.
func (m *Migrator) ListVersions(_ context.Context) ([]uint, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	seen := make(map[uint]struct{})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		if version := parseMigrationVersion(entry.Name()); version != 0 {
			seen[version] = struct{}{}
		}
	}

	versions := make([]uint, 0, len(seen))

	for version := range seen {
		versions = append(versions, version)
	}

	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })

	return versions, nil
}

func parseMigrationVersion(name string) uint {
	versionPart, _, ok := strings.Cut(name, "_")
	if !ok {
		return 0
	}

	version, err := strconv.ParseUint(versionPart, 10, 31)
	if err != nil {
		return 0
	}

	return uint(version)
}
