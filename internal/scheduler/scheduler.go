package scheduler

import (
	"fmt"
	"log/slog"
)

func New(config *Config) (*Scheduler, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil: %w", errInvalidInput)
	}

	if config.TaskProvider == nil {
		return nil, fmt.Errorf("task provider cannot be nil: %w", errInvalidConfig)
	}

	if config.Interval <= 0 {
		return nil, fmt.Errorf("interval cannot be zero or negative: %w", errInvalidConfig)
	}

	logger := slog.New(slog.DiscardHandler)
	if config.Logger != nil {
		logger = config.Logger
	}

	return &Scheduler{
		interval: config.Interval,
		logger:   logger,
		provider: config.TaskProvider,
		tasks:    make(map[string]*entry),
	}, nil
}
