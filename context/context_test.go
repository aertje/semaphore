package context

import (
	"context"
	"testing"

	"github.com/aertje/semaphore/semaphore"
	"github.com/stretchr/testify/assert"
)

func TestPrioritizedContext(t *testing.T) {
	t.Run("retrieves stored semaphore", func(t *testing.T) {
		ctx := context.Background()
		s := semaphore.NewPrioritized()

		ctx = WithPrioritized(ctx, s)
		retrieved, ok := PrioritizedFromContext(ctx)

		assert.True(t, ok)
		assert.Same(t, s, retrieved)
	})

	t.Run("returns false when no semaphore stored", func(t *testing.T) {
		ctx := context.Background()

		retrieved, ok := PrioritizedFromContext(ctx)

		assert.False(t, ok)
		assert.Nil(t, retrieved)
	})

	t.Run("handles nil semaphore", func(t *testing.T) {
		ctx := context.Background()

		ctx = WithPrioritized(ctx, nil)
		retrieved, ok := PrioritizedFromContext(ctx)

		assert.True(t, ok)
		assert.Nil(t, retrieved)
	})
}
