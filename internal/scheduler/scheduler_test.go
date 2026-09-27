package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNew(t *testing.T) {

	tests := []struct {
		name     string
		config   *Config
		want     *Scheduler
		checkErr func(t *testing.T, err error)
	}{
		{
			name: "good new scheduler",
			config: &Config{
				Interval: time.Second,
				Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
					Level: slog.LevelError,
				})),
				TaskProvider: &TestProvider{
					function: func(ctx context.Context) ([]Task, error) {
						return nil, nil
					},
				},
			},
			want: &Scheduler{
				interval: time.Second,
				tasks:    make(map[string]*entry),
			},
			checkErr: nil,
		},
		{
			name:   "nil config",
			config: nil,
			want:   nil,
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidInput)
				}

				if !errors.Is(err, errInvalidInput) {
					t.Fatalf("expected %v, got %v", errInvalidInput, err)
				}
			},
		},
		{
			name: "nil TaskProvider",
			config: &Config{
				Interval:     time.Second,
				Logger:       slog.New(slog.DiscardHandler),
				TaskProvider: nil,
			},
			want: nil,
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidConfig)
				}

				if !errors.Is(err, errInvalidConfig) {
					t.Fatalf("expected %v, got %v", errInvalidConfig, err)
				}
			},
		},
		{
			name: "invalid interval",
			config: &Config{
				Interval: 0,
				Logger:   slog.New(slog.DiscardHandler),
				TaskProvider: &TestProvider{
					function: func(ctx context.Context) ([]Task, error) {
						return nil, nil
					},
				},
			},
			want: nil,
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidConfig)
				}

				if !errors.Is(err, errInvalidConfig) {
					t.Fatalf("expected %v, got %v", errInvalidConfig, err)
				}
			},
		},
		{
			name: "nil logger",
			config: &Config{
				Interval: time.Millisecond,
				Logger:   nil,
				TaskProvider: &TestProvider{
					function: func(ctx context.Context) ([]Task, error) {
						return nil, nil
					},
				},
			},
			want: &Scheduler{
				interval: time.Millisecond,
				tasks:    make(map[string]*entry),
			},
			checkErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheduler, err := New(test.config)

			if test.checkErr != nil {
				test.checkErr(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.want != nil {
				if test.config == nil {
					t.Fatalf("config cannot be nil")
				}

				if test.config.Logger != nil {
					if scheduler.logger != test.config.Logger {
						t.Errorf("expected custom logger %p, got %p",
							test.config.Logger, scheduler.logger)
					}
				} else {
					if scheduler.logger == nil {
						t.Errorf("expected default logger to be initialized, got nil")
					}
				}

				if scheduler.provider != test.config.TaskProvider {
					t.Errorf("expected custom task provider %p, got %v",
						test.config.TaskProvider, scheduler.provider)
				}

				scheduler.logger.Info("test log to ensure logger initialization")
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(Scheduler{}),
				cmpopts.IgnoreTypes(sync.Mutex{}),
				cmpopts.IgnoreFields(Scheduler{}, "wg"),
				cmpopts.IgnoreFields(Scheduler{}, "logger"),
				cmpopts.IgnoreFields(Scheduler{}, "provider"),
			}
			if diff := cmp.Diff(test.want, scheduler, opts...); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
