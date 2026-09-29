# scheduler

An interval polling scheduler that uses the `TaskProvider` interface as its single source of truth.

Every interval, the scheduler asks the interface for a list of tasks. It then automatically updates itself to match that list: scheduling new tasks, updating existing ones if their start time changed and dropping tasks that are no longer provided.

## Quickstart

To run the scheduler and perform a recurring task:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "log/slog"
    "os"
    "os/signal"
    "time"

    "github.com/eepzii/f1trackside/internal/scheduler"
)

type Provider struct{}

func (p *Provider) Tasks(ctx context.Context) ([]scheduler.Task, error) {
    now := time.Now()

    return []scheduler.Task{
        {
            ID:        "teacher",
            StartTime: now,
            Do: func(ctx context.Context) {
                fmt.Println("Teacher: Good morning class!")
            },
        },
        {
            ID:        "student-alice",
            StartTime: now.Add(500 * time.Millisecond),
            Do: func(ctx context.Context) {
                fmt.Println("  Alice: Good morning!")
            },
        },
        {
            ID:        "student-bob",
            StartTime: now.Add(time.Second),
            Do: func(ctx context.Context) {
                fmt.Println("  Bob: Morning :)")
            },
        },
    }, ctx.Err()
}

func main() {
    provider := &Provider{}
    config := &scheduler.Config{
        Interval:     2 * time.Second,
        Logger:       slog.New(slog.DiscardHandler),
        TaskProvider: provider,
    }

    sched, err := scheduler.New(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
    defer cancel()

    sched.Run(ctx)
}
```

## Core Behavior

### State
* **Mirroring:** The scheduler always mirrors the latest successful `Tasks()` call from your `TaskProvider`.
* **Updates:** To change a task's start time, simply update what your `TaskProvider` returns. To change what a `Task` should do, remove the old one and add a new one with a new `ID`.

### Safety
* **Backpressure:** When a poll takes too long (e.g., due to a slow `TaskProvider` execution), the main loop will still run, but it will only trigger a new poll when the last poll has finished.
* **Panic Recovery:** If an individual task panics, it is caught and logged. One broken task will not crash the main scheduler.
* **Graceful Shutdown:** When the context is canceled, the scheduler stops polling and waits for all currently active tasks to finish before exiting.

## API Reference

### Methods

#### `func New(config *Config) (*Scheduler, error)`

Creates a new scheduler configured with an interval, logger and a `TaskProvider`.

#### `func (s *Scheduler) Run(ctx context.Context)`

Starts the main scheduling loop. It periodically queries the `TaskProvider` at the configured interval. If a poll takes too long, it drops ticks to handle backpressure. Canceling the provided context triggers a graceful shutdown, blocking until all currently running tasks complete.

### Types & Data Structures

#### `type Scheduler struct`

Holds the interval, `TaskProvider` and the internal state of all task entries.

#### `type Config struct`

Initialization requirements and optional parameters.

```go
type Config struct {
	Interval     time.Duration // required: must be greater than 0
	Logger       *slog.Logger  // optional: defaults to slog.New(slog.DiscardHandler)
	TaskProvider TaskProvider  // required: the single source of truth for tasks
}
```

#### `type Task struct`

Represents a single unit of work. The scheduler uses the `ID` to track the task and the `StartTime` to know when to schedule the `Do` function.

```go
type Task struct {
	ID        string
	StartTime time.Time

	Do func(ctx context.Context)
}
```

#### `type TaskProvider interface`

The interface that serves as the single source of truth for what the scheduler should be running.

```go
type TaskProvider interface {
	Tasks(ctx context.Context) ([]Task, error)
}
```