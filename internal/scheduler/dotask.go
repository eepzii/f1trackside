package scheduler

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"
)

func (s *Scheduler) doTask(ctx context.Context, task Task) func() {
	return func() {
		s.mu.Lock()
		if ctx.Err() != nil {
			s.mu.Unlock()
			return
		}

		entry, ok := s.tasks[task.ID]
		if !ok || !entry.startsAt(task.StartTime) {
			s.mu.Unlock()
			return
		}
		entry.isRunning = true
		s.wg.Add(1)
		s.mu.Unlock()

		func() {
			defer s.wg.Done()

			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("task panicked",
						"id", task.ID,
						slog.Any("panic", r),
						"stack", string(debug.Stack()),
					)
				}
			}()

			task.Do(ctx)
		}()

		s.mu.Lock()
		if entry, ok := s.tasks[task.ID]; ok {
			entry.isRunning = false

			if !entry.startsAt(task.StartTime) {
				duration := time.Until(entry.task.StartTime)
				entry.timer = time.AfterFunc(duration, s.doTask(ctx, entry.task))
			}
		}
		s.mu.Unlock()
	}
}
