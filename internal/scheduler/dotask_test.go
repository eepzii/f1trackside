package scheduler

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoTask_SkipsExecution(t *testing.T) {
	startTime := time.Now()

	tests := []struct {
		name   string
		action func(cancel context.CancelFunc, s *Scheduler, taskID string)
	}{
		{
			name: "context canceled",
			action: func(cancel context.CancelFunc, s *Scheduler, taskID string) {
				cancel()
			},
		},
		{
			name: "removed from map",
			action: func(_ context.CancelFunc, s *Scheduler, taskID string) {
				delete(s.tasks, taskID)
			},
		},
		{
			name: "start time changed",
			action: func(_ context.CancelFunc, s *Scheduler, taskID string) {
				s.tasks[taskID].task.StartTime = startTime.Add(time.Minute)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheduler := &Scheduler{
				logger: slog.New(slog.DiscardHandler),
				tasks:  make(map[string]*entry),
			}

			var executed atomic.Bool
			task := Task{
				ID:        "test-task",
				StartTime: startTime,
				Do: func(ctx context.Context) {
					executed.Store(true)
				},
			}

			scheduler.tasks[task.ID] = &entry{task: task}

			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()

			fn := scheduler.doTask(ctx, task)
			test.action(cancel, scheduler, task.ID)

			fn()

			if executed.Load() {
				t.Errorf("expected task to abort, but it executed")
			}

			if !scheduler.mu.TryLock() {
				t.Fatalf("expected mutex to be unlocked after abort")
			}
			scheduler.mu.Unlock()
		})
	}
}

func TestDoTask_IsRunning(t *testing.T) {
	scheduler := &Scheduler{
		logger: slog.New(slog.DiscardHandler),
		tasks:  make(map[string]*entry),
	}

	var isSet atomic.Bool
	task := Task{
		ID:        "test-task",
		StartTime: time.Now(),
		Do: func(ctx context.Context) {
			entry, ok := scheduler.tasks["test-task"]
			if !ok {
				return
			}

			if entry.isRunning {
				isSet.Store(true)
			}
		},
	}

	scheduler.tasks[task.ID] = &entry{task: task}
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	fn := scheduler.doTask(ctx, task)

	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal(".doTask() took too long, might be deadlocked")
	}

	if !isSet.Load() {
		t.Fatal(`expected "entry.isRunning = true" during task execution`)
	}

	if entry, ok := scheduler.tasks[task.ID]; ok {
		if entry.isRunning {
			t.Fatal(`expected "entry.isRunning = false" after function ends`)
		}
	}
}

func TestDoTask_Recovers(t *testing.T) {
	var logBuffer bytes.Buffer
	scheduler := &Scheduler{
		logger: slog.New(slog.NewTextHandler(&logBuffer, nil)),
		tasks:  make(map[string]*entry),
	}

	taskID := "test-task"
	task := Task{
		ID:        taskID,
		StartTime: time.Now(),
		Do: func(ctx context.Context) {
			panic("simulated failure")
		},
	}

	scheduler.tasks[task.ID] = &entry{task: task}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	fn := scheduler.doTask(ctx, task)

	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal(".doTask() took too long, might be deadlocked")
	}

	logOutput := logBuffer.String()
	expectedPhrases := []string{
		"task panicked",
		taskID,
		"simulated failure",
		"stack=",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(logOutput, phrase) {
			t.Errorf("expected log to contain %q, got: %q", phrase, logOutput)
		}
	}
}

func TestDoTask_Reschedules(t *testing.T) {

	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "fast 40ms",
			duration: 40 * time.Millisecond,
		},
		{
			name:     "mid 80ms",
			duration: 80 * time.Millisecond,
		},
		{
			name:     "slow 120ms",
			duration: 120 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheduler := &Scheduler{
				logger: slog.New(slog.DiscardHandler),
				tasks:  make(map[string]*entry),
			}

			var scheduleTime time.Time
			executed := make(chan time.Time)

			task := Task{
				ID:        "test-task",
				StartTime: time.Now(),
				Do: func(ctx context.Context) {
					scheduler.mu.Lock()
					scheduleTime = time.Now()

					scheduler.tasks["test-task"] = &entry{
						task: Task{
							ID:        "test-task",
							StartTime: scheduleTime.Add(test.duration),
							Do: func(ctx context.Context) {
								executed <- time.Now()
							},
						},
					}

					scheduler.mu.Unlock()
				},
			}

			scheduler.tasks[task.ID] = &entry{task: task}
			ctx, cancel := context.WithTimeout(t.Context(), time.Second)
			defer cancel()

			fn := scheduler.doTask(ctx, task)

			done := make(chan struct{})
			go func() {
				defer close(done)
				fn()
			}()

			select {
			case <-done:
			case <-time.After(100 * time.Millisecond):
				t.Fatal(".doTask() took too long, might be deadlocked")
			}

			scheduler.mu.Lock()
			entry, ok := scheduler.tasks[task.ID]
			if !ok || entry.timer == nil {
				scheduler.mu.Unlock()
				t.Fatalf("expected a new timer to be created for the rescheduled task")
			}

			timer := entry.timer
			scheduler.mu.Unlock()

			defer timer.Stop()

			select {
			case e := <-executed:
				diff := e.Sub(scheduleTime)
				buffer := 15 * time.Millisecond
				if diff < test.duration-buffer || diff > test.duration+buffer {
					t.Fatalf("rescheduled task executed in ~%v, want ~%v (±%v)",
						diff, test.duration, buffer)
				}
			case <-time.After(test.duration * 3):
				t.Fatalf("rescheduled task execution timed out after %v", test.duration*3)
			}
		})
	}
}

func TestDoTask_SkipsRescheduling(t *testing.T) {

	scheduler := &Scheduler{
		logger: slog.New(slog.DiscardHandler),
		tasks:  make(map[string]*entry),
	}

	startTime := time.Now()

	ctx, cancel := context.WithCancel(t.Context())

	task := Task{
		ID:        "test-task",
		StartTime: startTime,
		Do: func(ctx context.Context) {
			scheduler.mu.Lock()
			scheduler.tasks["test-task"] = &entry{
				task: Task{
					ID:        "test-task",
					StartTime: startTime.Add(time.Second),
					Do:        func(ctx context.Context) {},
				},
			}
			scheduler.mu.Unlock()

			cancel()
		},
	}

	scheduler.tasks[task.ID] = &entry{
		task: task,
	}

	fn := scheduler.doTask(ctx, task)

	done := make(chan struct{})
	go func() {
		defer close(done)
		fn()
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal(".doTask() took too long, might be deadlocked")
	}

	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()

	entry, ok := scheduler.tasks[task.ID]
	if !ok {
		t.Fatalf("expected task %q to remain in map", task.ID)
	}

	if entry.timer != nil {
		entry.timer.Stop()
		t.Fatal("expected no timer to be created after context cancellation")
	}
}

func TestDoTask_Concurrency(t *testing.T) {
	scheduler := &Scheduler{
		logger: slog.New(slog.DiscardHandler),
		tasks:  make(map[string]*entry),
	}

	routines := 20
	var executions atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := range routines {
		task := Task{
			ID:        fmt.Sprintf("%d-test-task", i),
			StartTime: time.Now(),
			Do: func(ctx context.Context) {
				executions.Add(1)
			},
		}

		scheduler.tasks[task.ID] = &entry{task: task}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		fn := scheduler.doTask(ctx, task)

		wg.Go(func() {
			<-start
			fn()
		})
	}

	close(start)

	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
		scheduler.wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("functions took too long, might be deadlocked")
	}

	if int(executions.Load()) != routines {
		t.Fatalf("expected %d executed tasks, got %d", routines, executions.Load())
	}
}
