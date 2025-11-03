package msync_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/graingo/maltose/util/msync"
	"github.com/stretchr/testify/assert"
)

func TestPool_Basic(t *testing.T) {
	t.Run("get_and_put", func(t *testing.T) {
		var createCount int32
		pool := msync.NewPool(
			10,
			func() any {
				atomic.AddInt32(&createCount, 1)
				return "object"
			},
			nil,
		)

		// Get an object
		obj := pool.Get()
		assert.Equal(t, "object", obj)
		assert.Equal(t, int32(1), createCount)

		// Put it back
		pool.Put(obj)

		// Get again, should reuse
		obj2 := pool.Get()
		assert.Equal(t, "object", obj2)
		assert.Equal(t, int32(1), createCount, "Should reuse object")
	})

	t.Run("put_nil_ignored", func(t *testing.T) {
		pool := msync.NewPool(5, func() any { return "object" }, nil)

		pool.Put(nil)
		assert.Equal(t, 0, pool.Available())
	})

	t.Run("multiple_objects", func(t *testing.T) {
		var createCount int32
		pool := msync.NewPool(
			5,
			func() any {
				count := atomic.AddInt32(&createCount, 1)
				return count
			},
			nil,
		)

		// Get 5 objects
		objects := make([]any, 5)
		for i := 0; i < 5; i++ {
			objects[i] = pool.Get()
		}

		assert.Equal(t, int32(5), createCount)
		assert.Equal(t, 5, pool.Size())

		// Put them back
		for _, obj := range objects {
			pool.Put(obj)
		}

		assert.Equal(t, 5, pool.Available())
	})
}

func TestPool_Limit(t *testing.T) {
	t.Run("respects_capacity_limit", func(t *testing.T) {
		const limit = 10
		var createCount int32

		pool := msync.NewPool(
			limit,
			func() any {
				return atomic.AddInt32(&createCount, 1)
			},
			nil,
		)

		// Get all objects
		objects := make([]any, limit)
		for i := 0; i < limit; i++ {
			objects[i] = pool.Get()
		}

		assert.Equal(t, limit, pool.Size())
		assert.Equal(t, int32(limit), createCount)

		// Put them back
		for _, obj := range objects {
			pool.Put(obj)
		}
	})

	t.Run("blocks_when_limit_reached", func(t *testing.T) {
		const limit = 3
		pool := msync.NewPool(
			limit,
			func() any { return "object" },
			nil,
		)

		// Get all objects
		for i := 0; i < limit; i++ {
			pool.Get()
		}

		// Try to get one more, should block
		done := make(chan bool)
		var obj any

		go func() {
			obj = pool.Get()
			done <- true
		}()

		// Should be blocked
		select {
		case <-done:
			t.Fatal("Should be blocked")
		case <-time.After(50 * time.Millisecond):
			// Good, it's blocked
		}

		// Put one back
		pool.Put("returned")

		// Now Get should succeed
		select {
		case <-done:
			assert.NotNil(t, obj)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Get should have succeeded")
		}
	})
}

func TestPool_Expiration(t *testing.T) {
	t.Run("expires_old_objects", func(t *testing.T) {
		var createCount, destroyCount int32

		pool := msync.NewPool(
			5,
			func() any {
				atomic.AddInt32(&createCount, 1)
				return "object"
			},
			func(any) {
				atomic.AddInt32(&destroyCount, 1)
			},
			msync.WithMaxAge(50*time.Millisecond),
		)

		// Get and put an object
		obj := pool.Get()
		pool.Put(obj)

		assert.Equal(t, int32(1), createCount)
		assert.Equal(t, int32(0), destroyCount)

		// Wait for expiration
		time.Sleep(100 * time.Millisecond)

		// Get again, should create new and destroy old
		obj2 := pool.Get()
		assert.NotNil(t, obj2)

		assert.Equal(t, int32(2), createCount, "Should create new object")
		assert.Equal(t, int32(1), destroyCount, "Should destroy expired object")
	})

	t.Run("skips_multiple_expired_objects", func(t *testing.T) {
		var createCount, destroyCount int32

		pool := msync.NewPool(
			10,
			func() any {
				return atomic.AddInt32(&createCount, 1)
			},
			func(any) {
				atomic.AddInt32(&destroyCount, 1)
			},
			msync.WithMaxAge(30*time.Millisecond),
		)

		// Create and return 5 objects
		objs := make([]any, 5)
		for i := 0; i < 5; i++ {
			objs[i] = pool.Get()
		}
		for i := 0; i < 5; i++ {
			pool.Put(objs[i])
		}

		assert.Equal(t, int32(5), atomic.LoadInt32(&createCount))

		// Wait for all to expire
		time.Sleep(50 * time.Millisecond)

		// Get should skip expired objects one by one
		pool.Get()

		// At least one expired object should be destroyed
		assert.Greater(t, atomic.LoadInt32(&destroyCount), int32(0))
	})
}

func TestPool_Destroy(t *testing.T) {
	t.Run("calls_destroy_on_expiration", func(t *testing.T) {
		destroyed := make(map[int]bool)
		var mu sync.Mutex

		pool := msync.NewPool(
			5,
			func() any { return new(int) },
			func(x any) {
				mu.Lock()
				destroyed[*x.(*int)] = true
				mu.Unlock()
			},
			msync.WithMaxAge(20*time.Millisecond),
		)

		// Create unique objects
		objs := make([]any, 3)
		for i := 0; i < 3; i++ {
			obj := pool.Get().(*int)
			*obj = i
			objs[i] = obj
		}
		for i := 0; i < 3; i++ {
			pool.Put(objs[i])
		}

		time.Sleep(50 * time.Millisecond)

		// Get should destroy at least one expired object
		pool.Get()

		mu.Lock()
		count := len(destroyed)
		mu.Unlock()

		assert.Greater(t, count, 0, "At least one expired object should be destroyed")
	})
}

func TestPool_Size(t *testing.T) {
	t.Run("tracks_total_objects", func(t *testing.T) {
		pool := msync.NewPool(10, func() any { return "object" }, nil)

		assert.Equal(t, 0, pool.Size())

		// Get 5 objects
		objects := make([]any, 5)
		for i := 0; i < 5; i++ {
			objects[i] = pool.Get()
		}

		assert.Equal(t, 5, pool.Size())

		// Put 3 back
		for i := 0; i < 3; i++ {
			pool.Put(objects[i])
		}

		// Size should still be 5 (total created)
		assert.Equal(t, 5, pool.Size())
	})
}

func TestPool_Available(t *testing.T) {
	t.Run("tracks_idle_objects", func(t *testing.T) {
		pool := msync.NewPool(10, func() any { return "object" }, nil)

		assert.Equal(t, 0, pool.Available())

		// Get and put
		obj1 := pool.Get()
		obj2 := pool.Get()

		pool.Put(obj1)
		assert.Equal(t, 1, pool.Available())

		pool.Put(obj2)
		assert.Equal(t, 2, pool.Available())

		// Get one
		pool.Get()
		assert.Equal(t, 1, pool.Available())
	})
}

func TestPool_Clear(t *testing.T) {
	t.Run("removes_all_idle_objects", func(t *testing.T) {
		var destroyCount int32

		pool := msync.NewPool(
			10,
			func() any { return "object" },
			func(any) { atomic.AddInt32(&destroyCount, 1) },
		)

		// Create and return 5 objects
		objs := make([]any, 5)
		for i := 0; i < 5; i++ {
			objs[i] = pool.Get()
		}
		for i := 0; i < 5; i++ {
			pool.Put(objs[i])
		}

		assert.Equal(t, 5, pool.Available())

		// Clear the pool
		pool.Clear()

		assert.Equal(t, 0, pool.Available())
		assert.Equal(t, 0, pool.Size())
		assert.Equal(t, int32(5), atomic.LoadInt32(&destroyCount))
	})

	t.Run("does_not_affect_in_use_objects", func(t *testing.T) {
		pool := msync.NewPool(10, func() any { return "object" }, nil)

		// Get 3 objects (in use)
		obj1 := pool.Get()
		obj2 := pool.Get()
		obj3 := pool.Get()

		// Put 2 back (idle)
		pool.Put(obj1)
		pool.Put(obj2)

		assert.Equal(t, 3, pool.Size())
		assert.Equal(t, 2, pool.Available())

		// Clear only removes idle objects
		pool.Clear()

		assert.Equal(t, 1, pool.Size(), "In-use object still counted")
		assert.Equal(t, 0, pool.Available())

		// Can still put obj3 back
		pool.Put(obj3)
		assert.Equal(t, 1, pool.Available())
	})
}

func TestPool_Concurrent(t *testing.T) {
	t.Run("concurrent_get_put", func(t *testing.T) {
		const goroutines = 100
		const operations = 1000

		pool := msync.NewPool(
			20,
			func() any { return new(int) },
			nil,
		)

		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					obj := pool.Get()
					// Simulate work
					pool.Put(obj)
				}
			}()
		}

		wg.Wait()

		// Pool should not exceed limit
		assert.LessOrEqual(t, pool.Size(), 20)
	})

	t.Run("stress_test", func(t *testing.T) {
		var createCount int32

		pool := msync.NewPool(
			50,
			func() any {
				atomic.AddInt32(&createCount, 1)
				return &struct{ data [1024]byte }{}
			},
			nil,
		)

		const workers = 100
		const iterations = 100

		var wg sync.WaitGroup
		wg.Add(workers)

		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					obj := pool.Get()
					time.Sleep(time.Microsecond)
					pool.Put(obj)
				}
			}()
		}

		wg.Wait()

		created := atomic.LoadInt32(&createCount)
		assert.LessOrEqual(t, created, int32(50), "Should not create more than limit")
	})
}

func TestPool_RealWorldScenarios(t *testing.T) {
	t.Run("database_connection_pool", func(t *testing.T) {
		type DBConn struct {
			ID     int
			Closed bool
		}

		var connID int32
		var closedCount int32

		pool := msync.NewPool(
			10,
			func() any {
				return &DBConn{
					ID: int(atomic.AddInt32(&connID, 1)),
				}
			},
			func(x any) {
				conn := x.(*DBConn)
				conn.Closed = true
				atomic.AddInt32(&closedCount, 1)
			},
			msync.WithMaxAge(5*time.Second),
		)

		// Simulate queries
		const queries = 50
		var wg sync.WaitGroup
		wg.Add(queries)

		for i := 0; i < queries; i++ {
			go func() {
				defer wg.Done()
				conn := pool.Get().(*DBConn)
				assert.False(t, conn.Closed)
				// Simulate query
				time.Sleep(time.Millisecond)
				pool.Put(conn)
			}()
		}

		wg.Wait()

		assert.LessOrEqual(t, pool.Size(), 10)
	})

	t.Run("buffer_pool", func(t *testing.T) {
		type Buffer struct {
			data []byte
		}

		pool := msync.NewPool(
			20,
			func() any {
				return &Buffer{data: make([]byte, 1024)}
			},
			func(x any) {
				// Clear buffer on destroy
				buf := x.(*Buffer)
				buf.data = nil
			},
		)

		// Simulate processing
		const tasks = 100
		var wg sync.WaitGroup
		wg.Add(tasks)

		for i := 0; i < tasks; i++ {
			go func(id int) {
				defer wg.Done()
				buf := pool.Get().(*Buffer)
				// Use buffer
				buf.data[0] = byte(id)
				pool.Put(buf)
			}(i)
		}

		wg.Wait()
	})
}

func BenchmarkPool_Get(b *testing.B) {
	b.Run("no_contention", func(b *testing.B) {
		pool := msync.NewPool(
			b.N,
			func() any { return &struct{}{} },
			nil,
		)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			obj := pool.Get()
			pool.Put(obj)
		}
	})

	b.Run("with_contention", func(b *testing.B) {
		pool := msync.NewPool(
			100,
			func() any { return &struct{}{} },
			nil,
		)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				obj := pool.Get()
				pool.Put(obj)
			}
		})
	})

	b.Run("with_expiration", func(b *testing.B) {
		pool := msync.NewPool(
			100,
			func() any { return &struct{}{} },
			nil,
			msync.WithMaxAge(1*time.Second),
		)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				obj := pool.Get()
				pool.Put(obj)
			}
		})
	})
}
