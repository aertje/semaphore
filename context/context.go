package context

import (
	"context"

	"github.com/aertje/semaphore/semaphore"
)

type prioritizedKey struct{}

var key = prioritizedKey{}

func PrioritizedFromContext(ctx context.Context) (*semaphore.Prioritized, bool) {
	s, ok := ctx.Value(key).(*semaphore.Prioritized)
	return s, ok
}

func WithPrioritized(ctx context.Context, s *semaphore.Prioritized) context.Context {
	return context.WithValue(ctx, key, s)
}
