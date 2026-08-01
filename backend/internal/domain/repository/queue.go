package repository

import (
	"context"
)

// MessageQueue is the message queue repository interface.
type MessageQueue interface {
	// Publish publishes a message to a queue.
	Publish(ctx context.Context, queue string, message []byte) error

	// Consume starts consuming from a queue with the given handler.
	Consume(ctx context.Context, queue string, handler MessageHandler) error

	// DeclareQueue declares a queue with options.
	DeclareQueue(ctx context.Context, name string, opts QueueOptions) error

	// Close shuts down the queue connection.
	Close(ctx context.Context) error
}

// MessageHandler processes a received message.
type MessageHandler func(ctx context.Context, message []byte) error

// QueueOptions defines queue declaration parameters.
type QueueOptions struct {
	Durable    bool
	AutoDelete bool
	Arguments  map[string]any
}
