package scheduler

import "context"

type TestProvider struct {
	function func(ctx context.Context) ([]Task, error)
}

func (m *TestProvider) Tasks(ctx context.Context) ([]Task, error) {
	return m.function(ctx)
}
