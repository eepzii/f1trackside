package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type TaskProvider interface {
	Tasks(ctx context.Context) ([]Task, error)
}

type Task struct {
	ID        string
	StartTime time.Time

	Do func(ctx context.Context)
}

type Config struct {
	Interval     time.Duration
	Logger       *slog.Logger
	TaskProvider TaskProvider
}

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
