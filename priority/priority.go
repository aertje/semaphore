package priority

import (
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

	movement chan struct{}
}

func New(maxConcurrency int) *P {
	p := &P{
		maxConcurrency: maxConcurrency,
		entries:        queue.New[entry](),
		movement:       make(chan struct{}),
	}
	p.schedule()
	return p
}

func (p *P) schedule() {
	go func() {
		for range p.movement {
			p.lock.Lock()
			if p.concurrency >= p.maxConcurrency {
				continue
			}

			entry, has := p.entries.Pop()
			if !has {
				continue
			}

			fnDone := func() {
				close(entry.waitChan)
				p.lock.Lock()
				defer p.lock.Unlock()
				p.concurrency--
				p.movement <- struct{}{}
			}

			entry.waitChan <- fnDone
			p.concurrency++

			p.lock.Unlock()
		}
	}()
}

func (p *P) Wait(priority int) chan func() {
	waitChan := make(chan func())

	entry := entry{
		priority: priority,
		waitChan: waitChan,
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	p.entries.Push(priority, entry)

	p.movement <- struct{}{}

	return waitChan
}
