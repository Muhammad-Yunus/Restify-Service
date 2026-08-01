package di

import "context"

// App is the application root with lifecycle control.
type App interface {
	// Run starts the application and blocks until ctx is cancelled.
	Run(ctx context.Context) error
	// Close releases application resources.
	Close(ctx context.Context) error
}

type app struct{}

func newApp() *app {
	return &app{}
}

// Run blocks until ctx is cancelled, then returns nil.
func (a *app) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (a *app) Close(ctx context.Context) error {
	return nil
}
