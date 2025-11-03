package msync_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/graingo/maltose/util/msync"
	"github.com/stretchr/testify/assert"
)

func TestLimit_Borrow(t *testing.T) {
	t.Run("basic_usage", func(t *testing.T) {
		limit := msync.NewLimit(5)

		// Should be able to borrow
		limit.Borrow()
		err := limit.Return()
		assert.NoError(t, err)
	})

	t.Run("blocks_when_full", func(t *testing.T) {
		limit := msync.NewLimit(2)

		// Borrow 2 slots
		limit.Borrow()
		limit.Borrow()

		// Third borrow should block
		borrowed := false
		done := make(chan bool)

		go func() {
			limit.Borrow()
			borrowed = true
			done <- true
		}()

		// Wait a bit to ensure the goroutine is blocked
		time.Sleep(50 * time.Millisecond)
		assert.False(t, borrowed, "Should be blocked waiting for slot")

		// Return one slot
		err := limit.Return()
		assert.NoError(t, err)

		// Now the third borrow should succeed
		select {
		case <-done:
			assert.True(t, borrowed)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Borrow should have succeeded after Return")
		}
	})

	t.Run("concurrent_limit_enforcement", func(t *testing.T) {
		const maxConcurrent = 10
		limit := msync.NewLimit(maxConcurrent)

		var currentConcurrent int32
		var maxReached int32

		const goroutines = 100
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()

				limit.Borrow()
				defer limit.Return()

				// Track concurrent executions
				current := atomic.AddInt32(&currentConcurrent, 1)

				// Update max if needed
				for {
					max := atomic.LoadInt32(&maxReached)
					if current <= max || atomic.CompareAndSwapInt32(&maxReached, max, current) {
						break
					}
				}

				// Simulate work
				time.Sleep(10 * time.Millisecond)

				atomic.AddInt32(&currentConcurrent, -1)
			}()
		}

		wg.Wait()

		// Should never exceed the limit
		max := atomic.LoadInt32(&maxReached)
		assert.LessOrEqual(t, max, int32(maxConcurrent),
			"Concurrent executions should not exceed limit")
	})
}

func TestLimit_TryBorrow(t *testing.T) {
	t.Run("succeeds_when_available", func(t *testing.T) {
		limit := msync.NewLimit(5)

		success := limit.TryBorrow()
		assert.True(t, success)

		err := limit.Return()
		assert.NoError(t, err)
	})

	t.Run("fails_when_full", func(t *testing.T) {
		limit := msync.NewLimit(2)

		// Fill the limit
		assert.True(t, limit.TryBorrow())
		assert.True(t, limit.TryBorrow())

		// Should fail now
		assert.False(t, limit.TryBorrow())

		// Return one
		err := limit.Return()
		assert.NoError(t, err)

		// Should succeed now
		assert.True(t, limit.TryBorrow())
	})

	t.Run("rate_limiting_scenario", func(t *testing.T) {
		const maxConcurrent = 5
		limit := msync.NewLimit(maxConcurrent)

		var accepted int32
		var rejected int32

		const requests = 20
		var wg sync.WaitGroup
		wg.Add(requests)

		for i := 0; i < requests; i++ {
			go func() {
				defer wg.Done()

				if limit.TryBorrow() {
					atomic.AddInt32(&accepted, 1)
					time.Sleep(10 * time.Millisecond)
					limit.Return()
				} else {
					atomic.AddInt32(&rejected, 1)
				}
			}()
		}

		wg.Wait()

		finalAccepted := atomic.LoadInt32(&accepted)
		finalRejected := atomic.LoadInt32(&rejected)

		assert.Equal(t, int32(requests), finalAccepted+finalRejected)
		assert.Greater(t, finalRejected, int32(0), "Some requests should be rejected")
	})
}

func TestLimit_Return(t *testing.T) {
	t.Run("error_on_extra_return", func(t *testing.T) {
		limit := msync.NewLimit(5)

		// Return without borrow should error
		err := limit.Return()
		assert.Error(t, err)
		assert.Equal(t, msync.ErrLimitReturn, err)
	})

	t.Run("correct_borrow_return_pairs", func(t *testing.T) {
		limit := msync.NewLimit(3)

		// Borrow and return should work
		limit.Borrow()
		err := limit.Return()
		assert.NoError(t, err)

		// Multiple pairs
		for i := 0; i < 10; i++ {
			limit.Borrow()
			err := limit.Return()
			assert.NoError(t, err)
		}
	})

	t.Run("defer_pattern", func(t *testing.T) {
		limit := msync.NewLimit(5)

		work := func() error {
			limit.Borrow()
			defer limit.Return()

			// Simulate work
			time.Sleep(5 * time.Millisecond)
			return nil
		}

		err := work()
		assert.NoError(t, err)
	})
}

func TestLimit_HTTPClientScenario(t *testing.T) {
	t.Run("batch_requests_with_limit", func(t *testing.T) {
		const maxConcurrent = 10
		limit := msync.NewLimit(maxConcurrent)

		const totalRequests = 100
		completed := int32(0)
		var wg sync.WaitGroup
		wg.Add(totalRequests)

		start := time.Now()

		for i := 0; i < totalRequests; i++ {
			go func(id int) {
				defer wg.Done()

				// Acquire slot
				limit.Borrow()
				defer limit.Return()

				// Simulate HTTP request
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&completed, 1)
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)

		assert.Equal(t, int32(totalRequests), completed)

		// With 100 requests at 10ms each and max 10 concurrent:
		// Minimum time = 100 / 10 * 10ms = 100ms
		assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond)
	})
}

func TestLimit_DatabaseConnectionPool(t *testing.T) {
	t.Run("simulate_connection_pool", func(t *testing.T) {
		const maxConnections = 5
		limit := msync.NewLimit(maxConnections)

		type Query struct {
			ID     int
			Result string
		}

		results := make(chan Query, 20)
		const queries = 20

		var wg sync.WaitGroup
		wg.Add(queries)

		for i := 0; i < queries; i++ {
			go func(queryID int) {
				defer wg.Done()

				// Wait for available connection
				limit.Borrow()
				defer limit.Return()

				// Simulate query execution
				time.Sleep(5 * time.Millisecond)

				results <- Query{
					ID:     queryID,
					Result: "success",
				}
			}(i)
		}

		wg.Wait()
		close(results)

		// Verify all queries completed
		count := 0
		for range results {
			count++
		}
		assert.Equal(t, queries, count)
	})
}

func BenchmarkLimit_Borrow(b *testing.B) {
	b.Run("no_contention", func(b *testing.B) {
		limit := msync.NewLimit(b.N)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			limit.Borrow()
			limit.Return()
		}
	})

	b.Run("high_contention", func(b *testing.B) {
		limit := msync.NewLimit(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				limit.Borrow()
				limit.Return()
			}
		})
	})
}

func BenchmarkLimit_TryBorrow(b *testing.B) {
	b.Run("always_available", func(b *testing.B) {
		limit := msync.NewLimit(b.N)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			if limit.TryBorrow() {
				limit.Return()
			}
		}
	})

	b.Run("high_contention", func(b *testing.B) {
		limit := msync.NewLimit(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if limit.TryBorrow() {
					limit.Return()
				}
			}
		})
	})
}
