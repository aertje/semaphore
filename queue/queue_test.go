package queue_test

import (
	"container/heap"
	"testing"

	"github.com/aertje/semaphore/queue"
	"github.com/stretchr/testify/assert"
)

func TestOrder(t *testing.T) {
	q := new(queue.Q[string])

	heap.Init(q)

	q.PushItem(2, "b")
	q.PushItem(3, "c")
	q.PushItem(1, "a")
	q.PushItem(2, "b")
	q.PushItem(1, "a")
	q.PushItem(3, "c")

	assert.Equal(t, 6, q.Len())

	for _, want := range []string{"a", "a", "b", "b", "c", "c", ""} {
		assert.Equal(t, want, q.PopItem())
	}

	assert.Equal(t, 0, q.Len())
}
