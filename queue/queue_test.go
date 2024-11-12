package queue_test

import (
	"container/heap"
	"testing"

	"github.com/aertje/gonice/queue"
	"github.com/stretchr/testify/assert"
)

func TestOrder(t *testing.T) {
	q := new(queue.Q[string])

	heap.Init(q)
	heap.Push(q, queue.NewItem(2, "b"))
	heap.Push(q, queue.NewItem(3, "c"))
	heap.Push(q, queue.NewItem(1, "a"))
	heap.Push(q, queue.NewItem(2, "b"))
	heap.Push(q, queue.NewItem(1, "a"))
	heap.Push(q, queue.NewItem(3, "c"))

	assert.Equal(t, 6, q.Len())

	for _, want := range []string{"a", "a", "b", "b", "c", "c"} {
		item := heap.Pop(q).(*queue.Item[string])
		assert.Equal(t, want, item.Value())
	}

	assert.Equal(t, 0, q.Len())
}
