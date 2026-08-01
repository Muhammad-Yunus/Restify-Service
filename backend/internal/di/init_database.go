package di

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
	"github.com/muhammadyunus/Restify-Service/internal/infrastructure/database"
)

func initDatabase(ctx context.Context, cfg config.DatabaseConfig) (repository.DB, *gorm.DB, error) {
	if cfg.URL == "" {
		return nil, nil, fmt.Errorf("database url is required")
	}

	pg, err := database.NewPostgresDB(ctx, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pg.Ping(ctx); err != nil {
		_ = pg.Close(ctx)
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	gormDB, err := database.NewGORMDB(pg)
	if err != nil {
		_ = pg.Close(ctx)
		return nil, nil, fmt.Errorf("create gorm: %w", err)
	}

	return pg, gormDB.DB(), nil
}
