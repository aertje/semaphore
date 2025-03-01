package queue

import "container/heap"

type item[T any] struct {
	value    T
	priority int
	index    int
}

type Q[T any] []*item[T]

// Len implements heap.Interface.
func (q Q[T]) Len() int {
	return len(q)
}

// Less implements heap.Interface.
func (q Q[T]) Less(i, j int) bool {
	return q[i].priority < q[j].priority
}

// Swap implements heap.Interface, do not use this method directly.
func (q Q[T]) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

// Push implements heap.Interface, do not use this method directly.
func (q *Q[T]) Push(x any) {
	n := len(*q)
	item := x.(*item[T])
	item.index = n
	*q = append(*q, item)
}

// Pop implements heap.Interface, do not use this method directly.
func (q *Q[T]) Pop() any {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*q = old[0 : n-1]
	return item
}

func (q *Q[T]) PushItem(priority int, value T) {
	item := &item[T]{value: value, priority: priority}
	heap.Push(q, item)
}

func (q *Q[T]) PopItem() T {
	if q.Len() == 0 {
		var zero T
		return zero
	}
	return heap.Pop(q).(*item[T]).value
}
