package queue

type Node[T any] struct {
	value    T
	priority int

	next *Node[T]
	prev *Node[T]
}

type Q[T any] struct {
	head *Node[T]
	tail *Node[T]
}

func New[T any]() *Q[T] {
	return &Q[T]{}
}

func (q *Q[T]) Push(priority int, value T) {
	node := &Node[T]{
		value:    value,
		priority: priority,
	}

	if q.head == nil {
		q.head = node
		q.tail = node
		return
	}

	if priority < q.head.priority {
		node.next = q.head
		q.head.prev = node
		q.head = node
		return
	}

	current := q.head

	for current.next != nil && priority >= current.next.priority {
		current = current.next
	}

	node.next = current.next
	node.prev = current
	current.next = node

	if node.next == nil {
		q.tail = node
	} else {
		node.next.prev = node
	}
}

func (q *Q[T]) Pop() (T, bool) {
	if q.head == nil {
		var zero T
		return zero, false
	}

	value := q.head.value

	if q.head == q.tail {
		q.head = nil
		q.tail = nil
		return value, true
	}

	q.head = q.head.next
	q.head.prev = nil

	return value, true
}
