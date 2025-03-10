package semaphore_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aertje/semaphore/semaphore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	s := semaphore.NewPrioritized()

	s.Acquire(1)
	s.Release()
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
	s := semaphore.NewPrioritized(semaphore.WithMaxConcurrency(maxConcurrency))

	// Saturate the scheduler otherwise subsequent tasks will be executed
	// immediately in undefined order.
	for range maxConcurrency {
		go func() {
			s.Acquire(0)
			defer s.Release()

			time.Sleep(10 * time.Millisecond)
		}()
	}

	// Give the scheduler some time to start the goroutines.
	time.Sleep(1 * time.Millisecond)

	results := make([]int, 0)
	var lock sync.Mutex
	var wg sync.WaitGroup

	for i := totalTasks / maxConcurrency; i > 0; i-- {
		for range maxConcurrency {
			priority := i
			wg.Add(1)
			go func() {
				defer wg.Done()

				s.Acquire(priority)
				defer s.Release()

				time.Sleep(10 * time.Millisecond)

				lock.Lock()
				defer lock.Unlock()
				results = append(results, i)
			}()
		}
	}

	wg.Wait()

	return results
}

func TestCancel(t *testing.T) {
	s := semaphore.NewPrioritized(semaphore.WithMaxConcurrency(1))

	// Saturate the scheduler otherwise the task under test will be executed
	// immediately without waiting.
	go func() {
		s.Acquire(0)
		defer s.Release()

		time.Sleep(10 * time.Millisecond)
	}()

	// Give the scheduler some time to start the goroutine.
	time.Sleep(1 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err := s.AcquireContext(ctx, 1)
	defer s.Release()

	require.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestForceAcquire(t *testing.T) {
	s := semaphore.NewPrioritized(semaphore.WithMaxConcurrency(1))

	// Saturate the scheduler otherwise subsequent tasks will be executed
	// immediately in undefined order.
	go func() {
		s.Acquire(0)
		defer s.Release()

		time.Sleep(10 * time.Millisecond)
	}()

	// Give the scheduler some time to start the goroutine.
	time.Sleep(1 * time.Millisecond)

	results := make([]bool, 0)
	var lock sync.Mutex
	var wg sync.WaitGroup

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			s.Acquire(0)
			defer s.Release()

			time.Sleep(10 * time.Millisecond)
			lock.Lock()
			defer lock.Unlock()
			results = append(results, false)
		}()
	}
	// Give the scheduler some time to start the goroutine.
	time.Sleep(1 * time.Millisecond)

	// These tasks start later, but they should finish earlier as they're
	// prioritized by the force acquire - while the normal prioritized tasks are
	// waiting for the concurrency to be available.
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			s.ForceAcquire()
			defer s.Release()

			lock.Lock()
			time.Sleep(10 * time.Millisecond)
			defer lock.Unlock()
			results = append(results, true)
		}()
	}

	wg.Wait()

	expected := []bool{true, true, true, true, true, false, false, false, false, false}
	require.Len(t, results, len(expected))
	assert.Equal(t, expected, results)
}
