# Epic 21 — Message Queue (RabbitMQ)

**Goal:** Implement RabbitMQ message queue for async task processing, email sending, and event bus.
**Dependencies:** Epic 06 (RabbitMQ adapter), Epic 05 (Queue repository interface)
**Commit:** `feat: add RabbitMQ message queue integration`

---

## Step 21.01 — Queue Service

**Build:** Create `backend/internal/application/service/queue_service.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// QueueService manages message queue operations.
type QueueService struct {
    queue repository.MessageQueue
}

// NewQueueService creates a new queue service.
func NewQueueService(queue repository.MessageQueue) *QueueService {
    return &QueueService{queue: queue}
}

// Publish publishes a message to a queue.
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

// DeclareQueue declares a queue with options.
func (qs *QueueService) DeclareQueue(ctx context.Context, name string, durable bool) error {
    return qs.queue.DeclareQueue(ctx, name, repository.QueueOptions{
        Durable:  durable,
        AutoDelete: false,
    })
}
```

---

## Step 21.02 — Worker Pattern

**Build:** Create `backend/internal/application/service/worker.go`:

```go
package service

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// Worker processes messages from a queue.
type Worker struct {
    name     string
    queue    string
    handler  repository.MessageHandler
    logger   repository.Logger
    consume  func(ctx context.Context, queue string, handler repository.MessageHandler) error
}

// NewWorker creates a new queue worker.
func NewWorker(name string, queue string, handler repository.MessageHandler, logger repository.Logger, consume func(ctx context.Context, queue string, handler repository.MessageHandler) error) *Worker {
    return &Worker{
        name:    name,
        queue:   queue,
        handler: handler,
        logger:  logger,
        consume: consume,
    }
}

// Start begins consuming messages.
func (w *Worker) Start(ctx context.Context) error {
    w.logger.Info(ctx, "starting worker", "name", w.name, "queue", w.queue)
    if err := w.consume(ctx, w.queue, w.handler); err != nil {
        return fmt.Errorf("worker %s: %w", w.name, err)
    }
    return nil
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

// StartAll starts all workers in goroutines.
func (wp *WorkerPool) StartAll(ctx context.Context) {
    for _, w := range wp.workers {
        go func(worker *Worker) {
            if err := worker.Start(ctx); err != nil {
                wp.logger.Error(ctx, "worker failed", "name", worker.name, "error", err)
            }
        }(w)
    }
}
```

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add RabbitMQ message queue with worker pool pattern"
```
