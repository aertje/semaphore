package context

import (
	"context"

	"github.com/aertje/semaphore/semaphore"
)

type schedulerKey struct{}

var key = schedulerKey{}

func SchedulerFromContext(ctx context.Context) (*semaphore.Prioritized, bool) {
	s, ok := ctx.Value(key).(*semaphore.Prioritized)
	return s, ok
}

func WithScheduler(ctx context.Context, s *semaphore.Prioritized) context.Context {
	return context.WithValue(ctx, key, s)
}
