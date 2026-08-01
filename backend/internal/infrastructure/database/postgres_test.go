package database

import (
	"context"
	"testing"

	"github.com/muhammadyunus/Restify-Service/internal/config"
)

func TestNewPostgresDB(t *testing.T) {
	ctx := context.Background()

	pg, err := NewPostgresDB(ctx, config.DatabaseConfig{URL: "postgres://user:pass@localhost:5432/db?sslmode=disable"})
	if err != nil {
		t.Fatalf("NewPostgresDB: %v", err)
	}

	if pg == nil {
		t.Fatal("PostgresDB is nil")
	}

	if pg.Pool() == nil {
		t.Error("pool is nil")
	}

	if pg.SQLDB() == nil {
		t.Error("sql db wrapper is nil")
	}

	if err := pg.Close(ctx); err != nil {
		t.Errorf("Close: %v", err)
	}
}

func TestNewPostgresDBInvalidDSN(t *testing.T) {
	_, err := NewPostgresDB(context.Background(), config.DatabaseConfig{URL: "not-a-valid-dsn"})
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}

func TestPostgresDBImplementsDBInterface(t *testing.T) {
	var _ interface {
		Ping(ctx context.Context) error
	} = (*PostgresDB)(nil)
}
