package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	"github.com/muhammadyunus/Restify-Service/internal/di"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load(filepath.Join("configs", ".env"))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	container, err := di.Bootstrap(ctx, cfg)
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	app, err := di.NewApp(container)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}
	defer app.Close(ctx)

	if err := app.Run(ctx); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
