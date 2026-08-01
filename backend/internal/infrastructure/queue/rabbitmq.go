package queue

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// RabbitMQQueue implements the repository.MessageQueue interface.
type RabbitMQQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewRabbitMQQueue creates a new RabbitMQ connection.
func NewRabbitMQQueue(ctx context.Context, url string) (*RabbitMQQueue, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()

		return nil, fmt.Errorf("create channel: %w", err)
	}

	return &RabbitMQQueue{conn: conn, channel: ch}, nil
}

// Publish publishes a message to a queue.
func (q *RabbitMQQueue) Publish(ctx context.Context, queue string, message []byte) error {
	if err := q.channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key = queue name
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
			MessageId:   uuid.New().String(),
		},
	); err != nil {
		return fmt.Errorf("publish to %s: %w", queue, err)
	}

	return nil
}

// Consume starts consuming from a queue with the given handler.
func (q *RabbitMQQueue) Consume(ctx context.Context, queue string, handler repository.MessageHandler) error {
	msgs, err := q.channel.ConsumeWithContext(
		ctx,
		queue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("consume %s: %w", queue, err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(ctx, msg.Body); err != nil {
				_ = msg.Nack(false, true) // requeue on error

				continue
			}

			_ = msg.Ack(false)
		}
	}()

	return nil
}

// DeclareQueue declares a queue with options.
func (q *RabbitMQQueue) DeclareQueue(ctx context.Context, name string, opts repository.QueueOptions) error {
	_, err := q.channel.QueueDeclare(
		name,
		opts.Durable,
		opts.AutoDelete,
		false,          // exclusive
		false,          // noWait
		opts.Arguments, // args
	)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", name, err)
	}

	return nil
}

// Close shuts down the queue connection.
func (q *RabbitMQQueue) Close(ctx context.Context) error {
	if err := q.channel.Close(); err != nil {
		return fmt.Errorf("close rabbitmq channel: %w", err)
	}

	if err := q.conn.Close(); err != nil {
		return fmt.Errorf("close rabbitmq connection: %w", err)
	}

	return nil
}

// Compile-time check.
var _ repository.MessageQueue = (*RabbitMQQueue)(nil)
