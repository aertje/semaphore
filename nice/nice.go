package nice

import (
	"context"
	"runtime"
	"sync"

	"github.com/aertje/gonice/queue"
)

type entry struct {
	priority   int
	waitChan   chan<- func()
	cancelChan <-chan struct{}
}

type Scheduler struct {
	maxConcurrency int

	concurrency int

	lock    sync.Mutex
	entries *queue.Q[entry]

	incoming chan entry
	done     chan entry
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
		entries:        queue.New[entry](),
		incoming:       make(chan entry),
		done:           make(chan entry),
	}

	for _, opt := range opts {
		opt(s)
	}

	s.schedule()
	return s
}

func (s *Scheduler) assessEntries() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for {
		if s.concurrency >= s.maxConcurrency {
			return
		}

		entry, has := s.entries.Pop()
		if !has {
			return
		}

		fnDone := func() {
			s.lock.Lock()
			s.concurrency--
			s.lock.Unlock()
			s.done <- entry
		}

		select {
		case <-entry.cancelChan:
			s.done <- entry
		default:
			entry.waitChan <- fnDone
			close(entry.waitChan)
			s.concurrency++
		}
	}
}

func (s *Scheduler) schedule() {
	go func() {
		for {
			select {
			case entry := <-s.incoming:
				s.lock.Lock()
				s.entries.Push(entry.priority, entry)
				s.lock.Unlock()
				s.assessEntries()
			case <-s.done:
				s.assessEntries()
			}
		}
	}()
}

func (s *Scheduler) WaitContext(ctx context.Context, priority int) func() {
	waitChan := make(chan func())
	cancelChan := make(chan struct{})

	entry := entry{
		priority:   priority,
		waitChan:   waitChan,
		cancelChan: cancelChan,
	}

	s.incoming <- entry

	select {
	case <-ctx.Done():
		close(cancelChan)
		return func() {}
	case fnDone := <-waitChan:
		return fnDone
	}
}

func (s *Scheduler) Wait(priority int) func() {
	return s.WaitContext(context.Background(), priority)
}
