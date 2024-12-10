package semaphore

import (
	"container/heap"
	"context"
	"runtime"
	"sync"

	"github.com/aertje/semaphore/queue"
)

type entry struct {
	waitChan   chan<- struct{}
	cancelChan <-chan struct{}
}

type Prioritized struct {
	maxConcurrency int

	concurrency int

	lock    sync.Mutex
	entries *queue.Q[entry]
}

type Option func(*Prioritized)

func WithMaxConcurrency(maxConcurrency int) Option {
	return func(p *Prioritized) {
		p.maxConcurrency = maxConcurrency
	}
}

func NewPrioritized(opts ...Option) *Prioritized {
	s := &Prioritized{
		maxConcurrency: runtime.GOMAXPROCS(0),
		entries:        new(queue.Q[entry]),
	}

	heap.Init(s.entries)

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Prioritized) assessEntries() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for {
		if s.concurrency >= s.maxConcurrency {
			return
		}

		if s.entries.Len() == 0 {
			return
		}
		entry := heap.Pop(s.entries).(*queue.Item[entry]).Value()

		select {
		case <-entry.cancelChan:
			continue
		default:
			entry.waitChan <- struct{}{}
			close(entry.waitChan)
			s.concurrency++
		}
	}
}

func (s *Prioritized) AcquireContext(ctx context.Context, priority int) error {
	waitChan := make(chan struct{})
	cancelChan := make(chan struct{})

	entry := entry{
		waitChan:   waitChan,
		cancelChan: cancelChan,
	}

	s.lock.Lock()
	heap.Push(s.entries, queue.NewItem(priority, entry))
	s.lock.Unlock()

	go func() {
		s.assessEntries()
	}()

	select {
	case <-ctx.Done():
		close(cancelChan)
		return ctx.Err()
	case <-waitChan:
		return nil
	}
}

func (s *Prioritized) Acquire(priority int) {
	s.AcquireContext(context.Background(), priority)
}

func (s *Prioritized) Release() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.concurrency--

	go func() {
		s.assessEntries()
	}()
}
