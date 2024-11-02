package nice

import (
	"runtime"
	"sync"

	"github.com/aertje/gonice/queue"
)

type entry struct {
	priority int
	waitChan chan func()
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

	if s.concurrency >= s.maxConcurrency {
		return
	}

	entry, has := s.entries.Pop()
	if !has {
		return
	}

	fnDone := func() {
		close(entry.waitChan)
		s.lock.Lock()
		s.concurrency--
		s.lock.Unlock()
		s.done <- entry
	}

	entry.waitChan <- fnDone
	s.concurrency++
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

func (s *Scheduler) Wait(priority int) chan func() {
	waitChan := make(chan func())

	entry := entry{
		priority: priority,
		waitChan: waitChan,
	}

	s.incoming <- entry

	return waitChan
}
