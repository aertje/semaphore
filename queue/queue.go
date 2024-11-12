package queue

type Item[T any] struct {
	value    T
	priority int
	index    int
}

func NewItem[T any](priority int, value T) *Item[T] {
	return &Item[T]{value: value, priority: priority}
}

func (item *Item[T]) Value() T {
	return item.value
}

type Q[T any] []*Item[T]

func (pq Q[T]) Len() int {
	return len(pq)
}

func (pq Q[T]) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}

func (pq Q[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *Q[T]) Push(x any) {
	n := len(*pq)
	item := x.(*Item[T])
	item.index = n
	*pq = append(*pq, item)
}

func (pq *Q[T]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
