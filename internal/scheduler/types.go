package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// TaskProvider is the single source of truth for the Scheduler.
type TaskProvider interface {
	Tasks(ctx context.Context) ([]Task, error)
}

// Task represents a single unit of work.
type Task struct {
	ID        string
	StartTime time.Time

	Do func(ctx context.Context)
}

// Config defines the initialization options for creating a new Scheduler.
type Config struct {
	Interval     time.Duration
	Logger       *slog.Logger
	TaskProvider TaskProvider
}

// Scheduler manages the interval polling, the lifecycle and concurrent execution of task entries.
type Scheduler struct {
	interval time.Duration
	logger   *slog.Logger
	provider TaskProvider

	mu    sync.Mutex
	wg    sync.WaitGroup
	tasks map[string]*entry
}

type entry struct {
	task      Task
	timer     *time.Timer
	isRunning bool
}

func (e *entry) startsAt(t time.Time) bool {
	return e.task.StartTime.Equal(t)
}
