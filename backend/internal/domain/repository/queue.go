package repository

import "context"

// MessageQueue is the message queue repository interface.
type MessageQueue interface {
	// Close shuts down the queue connection.
	Close(ctx context.Context) error
}
