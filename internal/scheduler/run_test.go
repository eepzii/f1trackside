package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestRun_Lifecycle(t *testing.T) {
	polled := make(chan struct{}, 1)

	scheduler := &Scheduler{
		logger: slog.New(slog.DiscardHandler),
		provider: &TestProvider{
			function: func(ctx context.Context) ([]Task, error) {
				select {
				case polled <- struct{}{}:
				default:
				}
				return nil, errors.New("simulated error")
			},
		},
		interval: time.Hour,
		tasks:    make(map[string]*entry),
	}

	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan struct{})
	go func() {
		defer close(done)
		scheduler.Run(ctx)
	}()

	select {
	case <-polled:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected initial poll to trigger immediately")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal(".Run() hung and did not shut down on context cancellation")
	}
}

func TestRun_PendingTasks(t *testing.T) {
	t.Run("stop timers", func(t *testing.T) {
		trap := make(chan struct{}, 1)

		scheduler := &Scheduler{
			interval: time.Hour,
			logger:   slog.New(slog.DiscardHandler),
			provider: &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					return nil, errors.New("simulated error")
				},
			},
			tasks: map[string]*entry{
				"test-task": {
					timer: time.AfterFunc(50*time.Millisecond, func() {
						close(trap)
					}),
				},
			},
		}

		ctx, cancel := context.WithCancel(t.Context())

		done := make(chan struct{})
		go func() {
			defer close(done)
			scheduler.Run(ctx)
		}()

		cancel()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal(".Run() did not exit promptly")
		}

		select {
		case <-trap:
			t.Fatal("timer was not stopped during scheduler shutdown")
		case <-time.After(75 * time.Millisecond):
		}
	})
}

func TestRun_Interval(t *testing.T) {

	tests := []struct {
		name     string
		interval time.Duration
	}{
		{
			name:     "fast 40ms",
			interval: 40 * time.Millisecond,
		},
		{
			name:     "slow 90ms",
			interval: 90 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ticks := make(chan time.Time, 5)

			provider := &TestProvider{
				function: func(ctx context.Context) ([]Task, error) {
					ticks <- time.Now()
					return nil, errors.New("simulated error")
				},
			}

			scheduler := &Scheduler{
				logger:   slog.New(slog.DiscardHandler),
				provider: provider,
				interval: test.interval,
				tasks:    make(map[string]*entry),
			}

			ctx, cancel := context.WithCancel(t.Context())
			go scheduler.Run(ctx)

			var timestamps []time.Time

			for i := range 3 {
				select {
				case stamp := <-ticks:
					timestamps = append(timestamps, stamp)
				case <-time.After(test.interval * 3):
					t.Fatalf("timed out waiting for tick %d", i)
				}
			}

			cancel()

			for i := 1; i < len(timestamps); i++ {
				diff := timestamps[i].Sub(timestamps[i-1])

				buffer := 15 * time.Millisecond

				if diff < test.interval-buffer || diff > test.interval+buffer {
					t.Errorf("expected interval ~%v, got gap of %v", test.interval, diff)
				}
			}
		})
	}
}

func TestRun_Backpressure(t *testing.T) {
	t.Run("probabilistic deadlock trap (25 runs = 99.99999% certainty)", func(t *testing.T) {
		for range 25 {
			freeze := make(chan struct{})
			started := make(chan struct{})

			scheduler := &Scheduler{
				interval: 5 * time.Millisecond,
				logger:   slog.New(slog.DiscardHandler),
				provider: &TestProvider{
					function: func(ctx context.Context) ([]Task, error) {
						select {
						case started <- struct{}{}:
						default:
						}

						<-freeze
						return nil, errors.New("simulated error")
					},
				},
				tasks: make(map[string]*entry),
			}

			ctx, cancel := context.WithCancel(t.Context())

			done := make(chan struct{})
			go func() {
				defer close(done)
				scheduler.Run(ctx)
			}()

			<-started

			time.Sleep(15 * time.Millisecond)

			cancel()
			close(freeze)

			select {
			case <-done:
			case <-time.After(100 * time.Millisecond):
				t.Fatalf(".Run() deadlocked on backpressure")
			}
		}
	})
}

func TestRun_ShutdownMutex(t *testing.T) {
	scheduler := &Scheduler{
		logger:   slog.New(slog.DiscardHandler),
		interval: time.Hour,
		provider: &TestProvider{
			function: func(ctx context.Context) ([]Task, error) {
				return nil, errors.New("simulated error")
			},
		},
		tasks: map[string]*entry{},
	}

	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan struct{})
	go func() {
		defer close(done)
		scheduler.Run(ctx)
	}()

	scheduler.mu.Lock()
	cancel()

	select {
	case <-done:
		t.Fatal(".Run() exited while mutex was held")
	case <-time.After(50 * time.Millisecond):
	}

	scheduler.mu.Unlock()

	select {
	case <-done:
	case <-time.After(50 * time.Millisecond):
		t.Fatal(".Run() hung after mutex was released")
	}

	if !scheduler.mu.TryLock() {
		t.Fatal(".Run() failed to unlock mutex on exit")
	}
	scheduler.mu.Unlock()
}
