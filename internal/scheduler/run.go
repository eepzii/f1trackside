package scheduler

import (
	"context"
	"sync"
	"time"
)

func (s *Scheduler) Run(ctx context.Context) {
	trigger := make(chan struct{}, 1)
	var wg sync.WaitGroup

	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				s.mu.Lock()
				for _, entry := range s.tasks {
					entry.timer.Stop()
				}
				s.mu.Unlock()
				return
			case <-trigger:
				s.poll(ctx)
			}
		}
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	trigger <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			s.logger.Info("graceful shutdown: waiting for active tasks")
			s.wg.Wait()
			s.logger.Info("graceful shutdown: complete")
			return
		case <-ticker.C:
			select {
			case trigger <- struct{}{}:
			default:
				s.logger.Warn("tick dropped: handle task backpressure")
			}
		}
	}
}
