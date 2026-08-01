package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/muhammadyunus/Restify-Service/internal/domain/repository"
)

func TestWorkerStart(t *testing.T) {
	var started bool
	worker := NewWorker(
		"test-worker",
		"test-queue",
		func(ctx context.Context, message []byte) error {
			return nil
		},
		&mockLogger{},
		func(ctx context.Context, queue string, handler repository.MessageHandler) error {
			started = true
			return nil
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go worker.Start(ctx)

	time.Sleep(50 * time.Millisecond)
	if !started {
		t.Fatal("expected worker to start")
	}
}

func TestWorkerStartFailure(t *testing.T) {
	expectedErr := errors.New("consume failed")
	logger := &mockLogger{}

	worker := NewWorker(
		"fail-worker",
		"test-queue",
		nil,
		logger,
		func(ctx context.Context, queue string, handler repository.MessageHandler) error {
			return expectedErr
		},
	)

	err := worker.Start(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestWorkerName(t *testing.T) {
	worker := NewWorker(
		"named-worker",
		"test-queue",
		nil,
		&mockLogger{},
		nil,
	)
	if worker.Name() != "named-worker" {
		t.Fatalf("expected name 'named-worker', got %s", worker.Name())
	}
}

func TestWorkerPoolAdd(t *testing.T) {
	pool := NewWorkerPool(&mockLogger{})

	worker1 := NewWorker("worker1", "queue1", nil, &mockLogger{}, nil)
	worker2 := NewWorker("worker2", "queue2", nil, &mockLogger{}, nil)

	pool.Add(worker1)
	pool.Add(worker2)

	if len(pool.Workers()) != 2 {
		t.Fatalf("expected 2 workers, got %d", len(pool.Workers()))
	}
	if pool.Workers()[0].Name() != "worker1" {
		t.Fatalf("expected first worker to be 'worker1', got %s", pool.Workers()[0].Name())
	}
	if pool.Workers()[1].Name() != "worker2" {
		t.Fatalf("expected second worker to be 'worker2', got %s", pool.Workers()[1].Name())
	}
}

func TestWorkerPoolStartAll(t *testing.T) {
	var worker1Started, worker2Started bool

	pool := NewWorkerPool(&mockLogger{})

	worker1 := NewWorker("worker1", "q1", nil, &mockLogger{}, func(ctx context.Context, queue string, handler repository.MessageHandler) error {
		worker1Started = true
		<-ctx.Done()
		return nil
	})

	worker2 := NewWorker("worker2", "q2", nil, &mockLogger{}, func(ctx context.Context, queue string, handler repository.MessageHandler) error {
		worker2Started = true
				<-ctx.Done()
		return nil
	})

	pool.Add(worker1)
	pool.Add(worker2)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go pool.StartAll(ctx)

	time.Sleep(100 * time.Millisecond)
	if !worker1Started {
		t.Fatal("expected worker1 to start")
	}
	if !worker2Started {
		t.Fatal("expected worker2 to start")
	}
}

func TestWorkerPoolStartSync(t *testing.T) {
	pool := NewWorkerPool(&mockLogger{})

	failingWorker := NewWorker("failing", "q1", nil, &mockLogger{}, func(ctx context.Context, queue string, handler repository.MessageHandler) error {
		return errors.New("simulated failure")
	})

	pool.Add(failingWorker)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	pool.StartSync(ctx)

	time.Sleep(100 * time.Millisecond)
	// Worker should have failed but pool should continue
}

func TestWorkerPoolLogger(t *testing.T) {
	logger := &mockLogger{}
	pool := NewWorkerPool(logger)

	if pool.Logger() != logger {
		t.Fatal("expected pool logger to match")
	}
}
