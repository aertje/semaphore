package priority_test

import (
	"sync"
	"testing"
	"time"

	"github.com/aertje/gonice/priority"
	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	p := priority.New()

	fnDone := <-p.Wait(1)
	fnDone()
}

func TestOrderSingleConcurrency(t *testing.T) {
	p := priority.New(priority.WithMaxConcurrency(1))

	waitChan := make(chan struct{})
	go func() {
		fnDone := <-p.Wait(0)
		waitChan <- struct{}{}
		time.Sleep(10 * time.Millisecond)
		fnDone()
	}()

	results := make([]int, 0)
	var lock sync.Mutex
	var wg sync.WaitGroup

	<-waitChan
	for i := 10; i > 0; i-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fnDone := <-p.Wait(i)
			time.Sleep(10 * time.Millisecond)
			lock.Lock()
			results = append(results, i)
			defer lock.Unlock()
			fnDone()
		}()
	}

	wg.Wait()

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, results)
}
