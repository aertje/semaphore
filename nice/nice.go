package nice

import (
	"container/heap"
	"context"
	"runtime"
	"sync"

	"github.com/aertje/gonice/queue"
)

type entry struct {
	priority   int
	waitChan   chan<- struct{}
	cancelChan <-chan struct{}
}

type Scheduler struct {
	maxConcurrency int

	concurrency int

	lock    sync.Mutex
	entries *queue.Q[entry]
}

type Option func(*Scheduler)

func WithMaxConcurrency(maxConcurrency int) Option {
	return func(p *Scheduler) {
		p.maxConcurrency = maxConcurrency
	}
}

func NewScheduler(opts ...Option) *Scheduler {
	s := &Scheduler{
		maxConcurrency: runtime.GOMAXPROCS(0),
		entries:        new(queue.Q[entry]),
	}

	heap.Init(s.entries)

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Scheduler) assessEntries() {
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

func (s *Scheduler) WaitContext(ctx context.Context, priority int) error {
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

func (s *Scheduler) Wait(priority int) {
	s.WaitContext(context.Background(), priority)
}

func (s *Scheduler) Done() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.concurrency--

	go func() {
		s.assessEntries()
	}()
}
