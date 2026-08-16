package scheduler

import (
	"context"
)

func (s *Scheduler) poll(ctx context.Context) {
	tasks, err := s.provider.Tasks(ctx)
	if err != nil {
		s.logger.Error("failed to fetch tasks", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	validIDs := s.registerTasks(ctx, tasks)

	for id, entry := range s.tasks {
		if _, ok := validIDs[id]; !ok {
			if entry.isRunning {
				continue
			}

			entry.timer.Stop()
			delete(s.tasks, id)
			s.logger.Debug("task unscheduled: removed from source", "id", id)
		}
	}
}
