package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/muhammadyunus/Restify-Service/internal/di"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app, err := di.Bootstrap(ctx)
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
	defer app.Close(ctx)

	if err := app.Run(ctx); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
