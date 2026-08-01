package di

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

// App is the top-level application object.
type App struct {
	container *Container
	server    *http.Server
	handler   http.Handler
}

// NewApp creates an App from a Container.
func NewApp(c *Container) (*App, error) {
	if c == nil {
		return nil, errors.New("nil container")
	}

	return &App{container: c, handler: c.Router}, nil
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func (a *App) Run(ctx context.Context) error {
	addr := net.JoinHostPort(a.container.Config.Server.Host, strconv.Itoa(a.container.Config.Server.Port))

	a.server = &http.Server{
		Addr:    addr,
		Handler: a.handler,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()

		_ = a.server.Shutdown(shutdownCtx)
	}()

	a.container.Logger.Info(ctx, "starting server", "addr", addr)

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

// Close shuts down all container dependencies.
func (a *App) Close(ctx context.Context) error {
	return a.container.Close(ctx)
}
