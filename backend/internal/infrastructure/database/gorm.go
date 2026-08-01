package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GORMDB wraps GORM sharing the pgx connection pool.
type GORMDB struct {
	db *gorm.DB
}

// NewGORMDB creates a GORM instance reusing the pgx pool connection.
func NewGORMDB(pg *PostgresDB) (*GORMDB, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: pg.SQLDB()}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	return &GORMDB{db: db}, nil
}

// DB returns the raw *gorm.DB for repository implementations.
func (g *GORMDB) DB() *gorm.DB {
	return g.db
}
