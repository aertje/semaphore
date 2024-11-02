package queue_test

import (
	"testing"

	"github.com/aertje/gonice/queue"
	"github.com/stretchr/testify/assert"
)

func TestOrder(t *testing.T) {
	q := queue.New[string]()

	q.Push(2, "b")
	q.Push(1, "a")
	q.Push(3, "c")

	for _, want := range []string{"a", "b", "c"} {
		val, has := q.Pop()
		assert.True(t, has)
		assert.Equal(t, want, val)
	}

	_, has := q.Pop()
	assert.False(t, has)
}
