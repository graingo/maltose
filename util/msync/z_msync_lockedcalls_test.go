package msync_test

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/graingo/maltose/util/msync"
	"github.com/stretchr/testify/assert"
)

func TestLockedCalls_Do(t *testing.T) {
	t.Run("basic_usage", func(t *testing.T) {
		lc := msync.NewLockedCalls()
		result, err := lc.Do("key1", func() (any, error) {
			return "value1", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "value1", result)
	})

	t.Run("each_call_executes_independently", func(t *testing.T) {
		lc := msync.NewLockedCalls()
		var callCount int32

		const goroutines = 10
		results := make([]any, goroutines)
		errors := make([]error, goroutines)
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(idx int) {
				defer wg.Done()
				results[idx], errors[idx] = lc.Do("same-key", func() (any, error) {
					count := atomic.AddInt32(&callCount, 1)
					time.Sleep(5 * time.Millisecond) // Simulate work
					return fmt.Sprintf("result-%d", count), nil
				})
			}(i)
		}

		wg.Wait()

		// All calls should execute
		assert.Equal(t, int32(goroutines), callCount)

		// Each should get a different result
		seen := make(map[string]bool)
		for i := 0; i < goroutines; i++ {
			assert.NoError(t, errors[i])
			result := results[i].(string)
			assert.False(t, seen[result], "Result should be unique: %s", result)
			seen[result] = true
		}
	})

	t.Run("sequential_execution_order", func(t *testing.T) {
		lc := msync.NewLockedCalls()
		var execution []int
		var mu sync.Mutex

		const goroutines = 5
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(idx int) {
				defer wg.Done()
				_, _ = lc.Do("ordered-key", func() (any, error) {
					mu.Lock()
					execution = append(execution, idx)
					mu.Unlock()
					time.Sleep(2 * time.Millisecond)
					return idx, nil
				})
			}(i)
		}

		wg.Wait()

		// All calls should execute
		assert.Equal(t, goroutines, len(execution))
	})

	t.Run("different_keys_execute_concurrently", func(t *testing.T) {
		lc := msync.NewLockedCalls()
		var callCount int32
		start := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				key := fmt.Sprintf("key-%d", idx)
				_, _ = lc.Do(key, func() (any, error) {
					atomic.AddInt32(&callCount, 1)
					time.Sleep(10 * time.Millisecond) // Each takes 10ms
					return idx, nil
				})
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(start)

		// All calls should execute
		assert.Equal(t, int32(10), callCount)

		// Should execute concurrently (much less than 10 * 10ms = 100ms)
		assert.Less(t, elapsed, 50*time.Millisecond, "Different keys should execute concurrently")
	})

	t.Run("error_handling", func(t *testing.T) {
		lc := msync.NewLockedCalls()
		expectedErr := errors.New("test error")

		result, err := lc.Do("error-key", func() (any, error) {
			return nil, expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("errors_are_independent", func(t *testing.T) {
		lc := msync.NewLockedCalls()
		var callCount int32

		const goroutines = 5
		errors := make([]error, goroutines)
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(idx int) {
				defer wg.Done()
				_, errors[idx] = lc.Do("key", func() (any, error) {
					count := atomic.AddInt32(&callCount, 1)
					if count == 1 {
						return nil, fmt.Errorf("error-%d", count)
					}
					return count, nil
				})
			}(i)
		}

		wg.Wait()

		// All calls should execute
		assert.Equal(t, int32(goroutines), callCount)

		// Only first call should error
		errorCount := 0
		for _, err := range errors {
			if err != nil {
				errorCount++
			}
		}
		assert.Equal(t, 1, errorCount, "Only one call should error")
	})

	t.Run("account_balance_update_scenario", func(t *testing.T) {
		// Simulate account balance updates that must be serialized
		lc := msync.NewLockedCalls()
		balance := int64(1000)

		const operations = 100
		var wg sync.WaitGroup
		wg.Add(operations)

		// 100 concurrent deposit operations
		for i := 0; i < operations; i++ {
			go func(amount int64) {
				defer wg.Done()
				_, _ = lc.Do("account-123", func() (any, error) {
					// Read current balance
					current := atomic.LoadInt64(&balance)
					time.Sleep(time.Microsecond) // Simulate processing
					// Update balance
					atomic.StoreInt64(&balance, current+amount)
					return current + amount, nil
				})
			}(10)
		}

		wg.Wait()

		// Final balance should be correct
		expected := int64(1000 + 100*10)
		assert.Equal(t, expected, atomic.LoadInt64(&balance))
	})

	t.Run("concurrent_file_writes", func(t *testing.T) {
		lc := msync.NewLockedCalls()
		writes := make(map[string][]int)
		var mu sync.Mutex

		const filesCount = 3
		const writesPerFile = 10
		var wg sync.WaitGroup

		for fileID := 0; fileID < filesCount; fileID++ {
			for writeID := 0; writeID < writesPerFile; writeID++ {
				wg.Add(1)
				go func(file, write int) {
					defer wg.Done()
					filename := fmt.Sprintf("file-%d", file)
					_, _ = lc.Do(filename, func() (any, error) {
						mu.Lock()
						writes[filename] = append(writes[filename], write)
						mu.Unlock()
						time.Sleep(time.Millisecond)
						return nil, nil
					})
				}(fileID, writeID)
			}
		}

		wg.Wait()

		// Each file should have all writes
		for i := 0; i < filesCount; i++ {
			filename := fmt.Sprintf("file-%d", i)
			assert.Equal(t, writesPerFile, len(writes[filename]),
				"File %s should have %d writes", filename, writesPerFile)
		}
	})
}

func BenchmarkLockedCalls_Do(b *testing.B) {
	b.Run("single_key", func(b *testing.B) {
		lc := msync.NewLockedCalls()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = lc.Do("benchmark-key", func() (any, error) {
					return "result", nil
				})
			}
		})
	})

	b.Run("multiple_keys", func(b *testing.B) {
		lc := msync.NewLockedCalls()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key-%d", i%10)
				_, _ = lc.Do(key, func() (any, error) {
					return i, nil
				})
				i++
			}
		})
	})

	b.Run("high_contention", func(b *testing.B) {
		lc := msync.NewLockedCalls()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = lc.Do("hotkey", func() (any, error) {
					// Simulate some work
					sum := 0
					for i := 0; i < 100; i++ {
						sum += i
					}
					return sum, nil
				})
			}
		})
	})
}
