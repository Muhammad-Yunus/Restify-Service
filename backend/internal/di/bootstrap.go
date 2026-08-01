package di

import "context"

// Bootstrap wires the application dependency graph and returns the root App.
func Bootstrap(ctx context.Context) (App, error) {
	return newApp(), nil
}
