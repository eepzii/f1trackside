package scheduler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPoll_ProviderError(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	provider := &TestProvider{
		function: func(ctx context.Context) ([]Task, error) {
			return nil, errors.New("simulated error")
		},
	}

	scheduler := &Scheduler{
		logger:   logger,
		provider: provider,
		tasks:    make(map[string]*entry),
	}

	entries := 20
	for i := range entries {
		taskID := fmt.Sprintf("%d-test-task", i)
		scheduler.tasks[taskID] = &entry{
			task: Task{
				ID: taskID,
			},
			timer: time.NewTimer(500 * time.Millisecond),
		}
	}

	defer func() {
		for _, e := range scheduler.tasks {
			if e.timer != nil {
				e.timer.Stop()
			}
		}
	}()

	scheduler.poll(t.Context())

	if len(scheduler.tasks) != 20 {
		t.Fatalf("expected poll to have %d map entries, got %d tasks", entries, len(scheduler.tasks))
	}

	logOutput := logBuffer.String()
	expectedPhrases := []string{
		"failed to fetch tasks",
		"error=",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(logOutput, phrase) {
			t.Errorf("expected log to contain %q, got: %q", phrase, logOutput)
		}
	}
}

func TestPoll_GuardClauses(t *testing.T) {
	startTime := time.Now().Add(time.Hour)
	dummyTask := func(ctx context.Context) {}

	tests := []struct {
		name       string
		provided   []Task
		checkTasks func(t *testing.T, tasks map[string]*entry)
	}{
		{
			name: "empty ID",
			provided: []Task{
				{
					ID:        "",
					StartTime: startTime,
					Do:        dummyTask,
				},
			},
			checkTasks: func(t *testing.T, tasks map[string]*entry) {
				if _, ok := tasks[""]; ok {
					t.Error("expected task with empty ID to be dropped")
				}
			},
		},
		{
			name: "nil .Do() function",
			provided: []Task{
				{
					ID:        "nil-test-task",
					StartTime: startTime,
					Do:        nil,
				},
			},
			checkTasks: func(t *testing.T, tasks map[string]*entry) {
				if _, ok := tasks["nil-test-task"]; ok {
					t.Error("expected task with nil .Do() function to be dropped")
				}
			},
		},
		{
			name: "duplicate IDs",
			provided: []Task{
				{
					ID:        "duplicate-test-task",
					StartTime: startTime,
					Do:        dummyTask,
				},
				{
					ID:        "duplicate-test-task",
					StartTime: time.Now().Add(30 * time.Minute),
					Do:        dummyTask,
				},
				{
					ID:        "duplicate-test-task",
					StartTime: time.Now().Add(45 * time.Minute),
					Do:        dummyTask,
				},
			},
			checkTasks: func(t *testing.T, tasks map[string]*entry) {
				e, ok := tasks["duplicate-test-task"]
				if !ok {
					t.Fatalf(`expected "duplicate" task to be in map`)
				}

				if !e.task.StartTime.Equal(startTime) {
					t.Errorf("expected first task to be included with .StartTime %q, got .StartTime %q",
						startTime, e.task.StartTime)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheduler := &Scheduler{
				logger: slog.New(slog.DiscardHandler),
				provider: &TestProvider{
					function: func(ctx context.Context) ([]Task, error) {
						return test.provided, nil
					},
				},
				tasks: make(map[string]*entry),
			}

			scheduler.poll(t.Context())

			defer func() {
				for _, e := range scheduler.tasks {
					if e.timer != nil {
						e.timer.Stop()
					}
				}
			}()

			test.checkTasks(t, scheduler.tasks)
		})
	}
}

func TestPoll_ExistingTasks(t *testing.T) {
	taskID := "test-task"
	baseTime := time.Now().Add(time.Hour)
	newTime := time.Now().Add(12 * time.Hour)
	dummyTask := func(ctx context.Context) {}

	t.Run("unchanged", func(t *testing.T) {
		timer := time.AfterFunc(time.Hour, func() {})

		scheduler := &Scheduler{
			logger: slog.New(slog.DiscardHandler),
			provider: &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					return []Task{
						{
							ID:        taskID,
							StartTime: baseTime,
							Do:        dummyTask,
						},
					}, nil
				},
			},
			tasks: map[string]*entry{
				taskID: {
					task: Task{
						ID:        taskID,
						StartTime: baseTime,
						Do:        dummyTask,
					},
					timer: timer,
				},
			},
		}

		scheduler.poll(t.Context())
		defer timer.Stop()

		if scheduler.tasks[taskID].timer != timer {
			t.Error("expected timer to remain the same")
		}
	})

	t.Run("running", func(t *testing.T) {
		timer := time.AfterFunc(time.Hour, func() {})

		scheduler := &Scheduler{
			logger: slog.New(slog.DiscardHandler),
			provider: &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					return []Task{
						{
							ID:        taskID,
							StartTime: newTime,
							Do:        dummyTask,
						},
					}, nil
				},
			},
			tasks: map[string]*entry{
				taskID: {
					task: Task{
						ID:        taskID,
						StartTime: baseTime,
						Do:        dummyTask,
					},
					timer:     timer,
					isRunning: true,
				},
			},
		}

		scheduler.poll(t.Context())
		defer timer.Stop()

		e := scheduler.tasks[taskID]
		if e.timer != timer {
			t.Error("expected timer to remain the same")
		}

		if !e.task.StartTime.Equal(newTime) {
			t.Errorf("expected .StartTime to update to %s, got %s",
				newTime, e.task.StartTime)
		}
	})

	t.Run("updated", func(t *testing.T) {
		trap := make(chan struct{}, 1)
		timer := time.AfterFunc(50*time.Millisecond, func() {
			close(trap)
		})

		scheduler := &Scheduler{
			logger: slog.New(slog.DiscardHandler),
			provider: &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					return []Task{
						{
							ID:        taskID,
							StartTime: newTime,
							Do:        dummyTask,
						},
					}, nil
				},
			},
			tasks: map[string]*entry{
				taskID: {
					task: Task{
						ID:        taskID,
						StartTime: baseTime,
						Do:        dummyTask,
					},
					timer: timer,
				},
			},
		}

		scheduler.poll(t.Context())

		defer func() {
			if e, ok := scheduler.tasks[taskID]; ok && e.timer != nil {
				e.timer.Stop()
			}
		}()

		if scheduler.tasks[taskID].timer == timer {
			t.Error("expected timer to be replaced")
		}

		select {
		case <-trap:
			t.Fatal("expected old timer to be stopped")
		case <-time.After(100 * time.Millisecond):
		}
	})
}

func TestPoll_TaskInitiation(t *testing.T) {

	tests := []struct {
		name  string
		delay time.Duration
	}{
		{
			name:  "short 40ms",
			delay: 40 * time.Millisecond,
		},
		{
			name:  "mid 80ms",
			delay: 80 * time.Millisecond,
		},
		{
			name:  "long 120ms",
			delay: 120 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			targetTime := time.Now()
			executed := make(chan time.Time)

			provider := &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					return []Task{
						{
							ID:        "test-task",
							StartTime: targetTime.Add(test.delay),
							Do: func(ctx context.Context) {
								executed <- time.Now()
							},
						},
					}, nil
				},
			}

			scheduler := &Scheduler{
				logger:   slog.New(slog.DiscardHandler),
				provider: provider,
				tasks:    make(map[string]*entry),
			}

			scheduler.poll(t.Context())

			defer func() {
				if e, ok := scheduler.tasks["test-task"]; ok && e.timer != nil {
					e.timer.Stop()
				}
			}()

			select {
			case e := <-executed:
				diff := e.Sub(targetTime)
				buffer := 15 * time.Millisecond
				if diff < test.delay-buffer || diff > test.delay+buffer {
					t.Fatalf("task executed in ~%v, want ~%v (±%v)", diff, test.delay, buffer)
				}
			case <-time.After(test.delay * 3):
				t.Fatalf("task execution timed out after %v", test.delay*3)
			}
		})
	}
}

func TestPoll_Deletion(t *testing.T) {
	taskID := "test-task"
	dummyTask := func(ctx context.Context) {}

	t.Run("running", func(t *testing.T) {
		timer := time.AfterFunc(time.Hour, func() {})

		scheduler := &Scheduler{
			logger: slog.New(slog.DiscardHandler),
			provider: &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					return []Task{}, nil
				},
			},
			tasks: map[string]*entry{
				taskID: {
					task: Task{
						ID: taskID,
						Do: dummyTask,
					},
					timer:     timer,
					isRunning: true,
				},
			},
		}

		scheduler.poll(t.Context())
		defer timer.Stop()

		if _, ok := scheduler.tasks[taskID]; !ok {
			t.Error("expected running task to remain in map")
		}
	})

	t.Run("deleted", func(t *testing.T) {
		trap := make(chan struct{}, 1)
		timer := time.AfterFunc(50*time.Millisecond, func() {
			close(trap)
		})

		scheduler := &Scheduler{
			logger: slog.New(slog.DiscardHandler),
			provider: &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					return []Task{}, nil
				},
			},
			tasks: map[string]*entry{
				taskID: {
					task: Task{
						ID: taskID,
						Do: dummyTask,
					},
					timer: timer,
				},
			},
		}

		scheduler.poll(t.Context())

		defer func() {
			if e, ok := scheduler.tasks[taskID]; ok && e.timer != nil {
				e.timer.Stop()
			}
		}()

		if _, ok := scheduler.tasks[taskID]; ok {
			t.Error("expected task to be deleted from map")
		}

		select {
		case <-trap:
			t.Fatal("expected timer to be stopped on task deletion")
		case <-time.After(100 * time.Millisecond):
		}
	})
}

func TestPoll_Concurrency(t *testing.T) {
	tasks := []Task{}
	dummyTask := func(ctx context.Context) {}

	scheduler := &Scheduler{
		logger: slog.New(slog.DiscardHandler),
		tasks: map[string]*entry{
			"0-delete-test-task": {
				timer: time.AfterFunc(time.Hour, func() {}),
			},
			"1-delete-test-task": {
				timer: time.AfterFunc(time.Hour, func() {}),
			},
			"2-delete-test-task": {
				timer: time.AfterFunc(time.Hour, func() {}),
			},
		},
	}

	routines := 20
	start := make(chan struct{})
	var wg sync.WaitGroup

	for i := range routines {
		tasks = append(tasks, Task{
			ID:        fmt.Sprintf("%d-test-task", i),
			StartTime: time.Now().Add(time.Hour),
			Do:        dummyTask,
		})
	}

	scheduler.provider = &TestProvider{
		function: func(ctx context.Context) ([]Task, error) {
			return tasks, nil
		},
	}

	for range routines {
		wg.Go(func() {
			<-start

			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()

			scheduler.poll(ctx)
		})
	}

	close(start)
	wg.Wait()

	if len(scheduler.tasks) != routines {
		t.Fatalf("expected exactly %d tasks in map, got %d", routines, len(scheduler.tasks))
	}

	for _, e := range scheduler.tasks {
		if e.timer != nil {
			e.timer.Stop()
		}
	}
}
