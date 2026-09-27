package scheduler

import (
	"context"
	"time"
)

func (s *Scheduler) poll(ctx context.Context) {
	provided, err := s.provider.Tasks(ctx)
	if err != nil {
		s.logger.Error("failed to fetch tasks", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	providedIDs := make(map[string]struct{})
	for _, task := range provided {
		if task.ID == "" {
			s.logger.Debug("task dropped: empty id")
			continue
		}

		if _, ok := providedIDs[task.ID]; ok {
			s.logger.Debug("task dropped: duplicate in source", "id", task.ID)
			continue
		}

		if task.Do == nil {
			s.logger.Debug("task dropped: missing required Do function", "id", task.ID)
			continue
		}
		providedIDs[task.ID] = struct{}{}

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

	for id, entry := range s.tasks {
		if _, ok := providedIDs[id]; !ok {
			if entry.isRunning {
				continue
			}

			entry.timer.Stop()
			delete(s.tasks, id)
			s.logger.Debug("task unscheduled: removed from source", "id", id)
		}
	}
}
