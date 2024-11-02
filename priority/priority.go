package priority

import (
	"runtime"
	"sync"

	"github.com/aertje/gonice/queue"
)

type entry struct {
	priority int
	waitChan chan func()
}

type P struct {
	maxConcurrency int

	concurrency int

	lock    sync.Mutex
	entries *queue.Q[entry]

	incoming chan entry
	done     chan entry
}

type Option func(*P)

func WithMaxConcurrency(maxConcurrency int) Option {
	return func(p *P) {
		p.maxConcurrency = maxConcurrency
	}
}

func New(opts ...Option) *P {
	p := &P{
		maxConcurrency: runtime.GOMAXPROCS(0),
		entries:        queue.New[entry](),
		incoming:       make(chan entry),
		done:           make(chan entry),
	}

	for _, opt := range opts {
		opt(p)
	}

	p.schedule()
	return p
}

func (p *P) assessEntries() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.concurrency >= p.maxConcurrency {
		return
	}

	entry, has := p.entries.Pop()
	if !has {
		return
	}

	fnDone := func() {
		close(entry.waitChan)
		p.lock.Lock()
		p.concurrency--
		p.lock.Unlock()
		p.done <- entry
	}

	entry.waitChan <- fnDone
	p.concurrency++
}

func (p *P) schedule() {
	go func() {
		for {
			select {
			case entry := <-p.incoming:
				p.lock.Lock()
				p.entries.Push(entry.priority, entry)
				p.lock.Unlock()
				p.assessEntries()
			case <-p.done:
				p.assessEntries()
			}
		}
	}()
}

func (p *P) Wait(priority int) chan func() {
	waitChan := make(chan func())

	entry := entry{
		priority: priority,
		waitChan: waitChan,
	}

	// p.lock.Lock()
	// defer p.lock.Unlock()
	// p.entries.Push(priority, entry)

	// p.movement <- struct{}{}

	p.incoming <- entry

	return waitChan
}
