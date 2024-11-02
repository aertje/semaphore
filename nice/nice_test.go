package nice_test

import (
	"sync"
	"testing"
	"time"

	"github.com/aertje/gonice/nice"
	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	s := nice.NewScheduler()

	fnDone := s.Wait(1)
	fnDone()
}

func TestOrderConcurrency(t *testing.T) {
	for _, tc := range []struct {
		name           string
		maxConcurrency int
		totalTasks     int
		expectedResult []int
	}{
		{
			name:           "no concurrency",
			maxConcurrency: 1,
			totalTasks:     10,
			expectedResult: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:           "concurrency 2",
			maxConcurrency: 2,
			totalTasks:     10,
			expectedResult: []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5},
		},
		{
			name:           "concurrency 8",
			maxConcurrency: 8,
			totalTasks:     16,
			expectedResult: []int{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			results := testOrderForConcurrency(tc.maxConcurrency, tc.totalTasks)
			assert.Equal(t, tc.expectedResult, results)
		})
	}
}

func testOrderForConcurrency(maxConcurrency int, totalTasks int) []int {
	s := nice.NewScheduler(nice.WithMaxConcurrency(maxConcurrency))

	// Saturate the scheduler otherwise subsequent tasks will be executed
	// immediately in undefined order.
	for i := 0; i < maxConcurrency; i++ {
		go func() {
			fnDone := s.Wait(0)
			time.Sleep(10 * time.Millisecond)
			fnDone()
		}()
	}

	// Give the scheduler some time to start the goroutines.
	time.Sleep(1 * time.Millisecond)

	results := make([]int, 0)
	var lock sync.Mutex
	var wg sync.WaitGroup

	for i := totalTasks / maxConcurrency; i > 0; i-- {
		for j := 0; j < maxConcurrency; j++ {
			priority := i
			wg.Add(1)
			go func() {
				defer wg.Done()

				fnDone := s.Wait(priority)

				time.Sleep(10 * time.Millisecond)

				lock.Lock()
				results = append(results, i)
				defer lock.Unlock()

				fnDone()
			}()
		}
	}

	wg.Wait()

	return results
}
