package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// QueueService manages message queue operations.
type QueueService struct {
	queue repository.MessageQueue
}

// NewQueueService creates a new queue service.
func NewQueueService(queue repository.MessageQueue) *QueueService {
	return &QueueService{queue: queue}
}

// Publish serializes and publishes a message to a queue.
func (qs *QueueService) Publish(ctx context.Context, queue string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	return qs.queue.Publish(ctx, queue, data)
}

// Consume starts consuming from a queue.
func (qs *QueueService) Consume(ctx context.Context, queue string, handler repository.MessageHandler) error {
	return qs.queue.Consume(ctx, queue, handler)
}

// DeclareQueue declares a durable queue.
func (qs *QueueService) DeclareQueue(ctx context.Context, name string, durable bool) error {
	return qs.queue.DeclareQueue(ctx, name, repository.QueueOptions{
		Durable:    durable,
		AutoDelete: false,
	})
}

// DeclareAndConsume is a convenience method to declare a queue and start consuming.
func (qs *QueueService) DeclareAndConsume(ctx context.Context, queue string, durable bool, handler repository.MessageHandler) error {
	if err := qs.DeclareQueue(ctx, queue, durable); err != nil {
		return fmt.Errorf("declare queue %s: %w", queue, err)
	}
	return qs.Consume(ctx, queue, handler)
}

// Close shuts down the queue connection.
func (qs *QueueService) Close(ctx context.Context) error {
	return qs.queue.Close(ctx)
}
