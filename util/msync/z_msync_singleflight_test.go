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

func TestSingleFlight_Do(t *testing.T) {
	t.Run("basic_usage", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		result, err := sf.Do("key1", func() (any, error) {
			return "value1", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "value1", result)
	})

	t.Run("concurrent_requests_share_result", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		var callCount int32

		const goroutines = 100
		results := make([]any, goroutines)
		errors := make([]error, goroutines)
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(idx int) {
				defer wg.Done()
				results[idx], errors[idx] = sf.Do("shared-key", func() (any, error) {
					atomic.AddInt32(&callCount, 1)
					time.Sleep(10 * time.Millisecond) // Simulate work
					return "shared-result", nil
				})
			}(i)
		}

		wg.Wait()

		// Should only execute once
		assert.Equal(t, int32(1), callCount)

		// All goroutines should get the same result
		for i := 0; i < goroutines; i++ {
			assert.NoError(t, errors[i])
			assert.Equal(t, "shared-result", results[i])
		}
	})

	t.Run("different_keys_execute_independently", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		var callCount int32

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				key := fmt.Sprintf("key-%d", idx)
				_, _ = sf.Do(key, func() (any, error) {
					atomic.AddInt32(&callCount, 1)
					return idx, nil
				})
			}(i)
		}

		wg.Wait()

		// Each key should execute once
		assert.Equal(t, int32(10), callCount)
	})

	t.Run("error_handling", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		expectedErr := errors.New("test error")

		result, err := sf.Do("error-key", func() (any, error) {
			return nil, expectedErr
		})

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, result)
	})

	t.Run("errors_are_shared", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		expectedErr := errors.New("shared error")

		const goroutines = 50
		errors := make([]error, goroutines)
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(idx int) {
				defer wg.Done()
				_, errors[idx] = sf.Do("error-key", func() (any, error) {
					time.Sleep(5 * time.Millisecond)
					return nil, expectedErr
				})
			}(i)
		}

		wg.Wait()

		// All goroutines should get the same error
		for i := 0; i < goroutines; i++ {
			assert.Equal(t, expectedErr, errors[i])
		}
	})

	t.Run("sequential_calls", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		var callCount int32

		// First call
		result1, err1 := sf.Do("key", func() (any, error) {
			atomic.AddInt32(&callCount, 1)
			return "first", nil
		})

		// Second call (after first completes)
		result2, err2 := sf.Do("key", func() (any, error) {
			atomic.AddInt32(&callCount, 1)
			return "second", nil
		})

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "first", result1)
		assert.Equal(t, "second", result2)
		assert.Equal(t, int32(2), callCount) // Should execute twice
	})
}

func TestSingleFlight_DoEx(t *testing.T) {
	t.Run("fresh_result", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		val, fresh, err := sf.DoEx("key", func() (any, error) {
			return "value", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "value", val)
		assert.True(t, fresh, "First call should return fresh result")
	})

	t.Run("shared_result", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		const goroutines = 50

		var freshCount int32
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				_, fresh, _ := sf.DoEx("shared-key", func() (any, error) {
					time.Sleep(10 * time.Millisecond)
					return "result", nil
				})
				if fresh {
					atomic.AddInt32(&freshCount, 1)
				}
			}()
		}

		wg.Wait()

		// Only one goroutine should get fresh result
		assert.Equal(t, int32(1), freshCount)
	})

	t.Run("distinguish_fresh_and_shared", func(t *testing.T) {
		sf := msync.NewSingleFlight()
		results := make(map[bool]int)
		var mu sync.Mutex

		const goroutines = 100
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				_, fresh, _ := sf.DoEx("key", func() (any, error) {
					time.Sleep(5 * time.Millisecond)
					return "value", nil
				})

				mu.Lock()
				results[fresh]++
				mu.Unlock()
			}()
		}

		wg.Wait()

		assert.Equal(t, 1, results[true], "Should have 1 fresh result")
		assert.Equal(t, goroutines-1, results[false], "Should have 99 shared results")
	})
}

func BenchmarkSingleFlight_Do(b *testing.B) {
	b.Run("single_key", func(b *testing.B) {
		sf := msync.NewSingleFlight()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = sf.Do("benchmark-key", func() (any, error) {
					time.Sleep(1 * time.Millisecond)
					return "result", nil
				})
			}
		})
	})

	b.Run("multiple_keys", func(b *testing.B) {
		sf := msync.NewSingleFlight()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key-%d", i%10)
				_, _ = sf.Do(key, func() (any, error) {
					return i, nil
				})
				i++
			}
		})
	})
}
