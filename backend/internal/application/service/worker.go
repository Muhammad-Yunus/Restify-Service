package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

// Worker processes messages from a queue.
type Worker struct {
	name    string
	queue   string
	handler repository.MessageHandler
	logger  repository.Logger
	consume func(ctx context.Context, queue string, handler repository.MessageHandler) error
}

// NewWorker creates a new queue worker.
func NewWorker(
	name string,
	queue string,
	handler repository.MessageHandler,
	logger repository.Logger,
	consume func(ctx context.Context, queue string, handler repository.MessageHandler) error,
) *Worker {
	return &Worker{
		name:    name,
		queue:   queue,
		handler: handler,
		logger:  logger,
		consume: consume,
	}
}

// Start begins consuming messages from the queue.
func (w *Worker) Start(ctx context.Context) error {
	w.logger.Info(ctx, "starting worker", "name", w.name, "queue", w.queue)
	if err := w.consume(ctx, w.queue, w.handler); err != nil {
		return fmt.Errorf("worker %s: %w", w.name, err)
	}

	// Block until context is cancelled
	<-ctx.Done()
	w.logger.Info(ctx, "worker stopped", "name", w.name, "queue", w.queue)
	return nil
}

// Name returns the worker name.
func (w *Worker) Name() string {
	return w.name
}

// WorkerPool manages multiple workers.
type WorkerPool struct {
	workers []*Worker
	logger  repository.Logger
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(logger repository.Logger) *WorkerPool {
	return &WorkerPool{logger: logger}
}

// Add adds a worker to the pool.
func (wp *WorkerPool) Add(w *Worker) {
	wp.workers = append(wp.workers, w)
}

// StartAll starts all workers in goroutines and returns a wait group.
func (wp *WorkerPool) StartAll(ctx context.Context) {
	for _, w := range wp.workers {
		go wp.runWorker(ctx, w)
	}
}

// StartSync starts all workers sequentially in goroutines.
func (wp *WorkerPool) StartSync(ctx context.Context) {
	for _, w := range wp.workers {
		go func(worker *Worker) {
			wp.logger.Info(ctx, "starting worker pool member", "name", worker.name)
			if err := worker.Start(ctx); err != nil {
				wp.logger.Error(ctx, "worker failed", "name", worker.name, "error", err)
			}
		}(w)
	}
}

func (wp *WorkerPool) runWorker(ctx context.Context, worker *Worker) {
	if err := worker.Start(ctx); err != nil {
		slog.Log(ctx, slog.LevelError, "worker pool failure", "worker", worker.name, "error", err)
	}
}

// Logger returns the pool logger.
func (wp *WorkerPool) Logger() repository.Logger {
	return wp.logger
}

// Workers returns the list of workers.
func (wp *WorkerPool) Workers() []*Worker {
	return wp.workers
}
