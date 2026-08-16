package scheduler

import (
	"context"
	"time"
)

func (s *Scheduler) registerTasks(ctx context.Context, provided []Task) map[string]struct{} {
	seen := make(map[string]struct{})

	for _, task := range provided {
		if task.ID == "" {
			s.logger.Debug("task dropped: empty id")
			continue
		}

		if _, ok := seen[task.ID]; ok {
			s.logger.Debug("task dropped: duplicate in source", "id", task.ID)
			continue
		}
		seen[task.ID] = struct{}{}

		if entry, ok := s.tasks[task.ID]; ok {
			if entry.startsAt(task.StartTime) {
				continue
			}

			if entry.isRunning {
				entry.task = task
				continue
			}

			s.logger.Debug("updating task start time",
				"id", task.ID,
				"old", entry.task.StartTime.UTC(),
				"new", task.StartTime.UTC(),
			)

			entry.timer.Stop()
		}

		duration := time.Until(task.StartTime)
		timer := time.AfterFunc(duration, s.doTask(ctx, task))

		s.tasks[task.ID] = &entry{
			task:  task,
			timer: timer,
		}
	}

	return seen
}
