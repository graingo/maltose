// Package msync provides concurrency control utilities for managing
// concurrent operations in Go applications.
//
// The package includes the following components:
//
//   - SingleFlight: Prevents duplicate function calls for the same key,
//     useful for preventing cache stampede.
//
//   - LockedCalls: Ensures sequential execution of operations with the same key,
//     useful for write operations that must be serialized.
//
//   - Limit: Controls the maximum number of concurrent operations,
//     useful for rate limiting and resource management.
//
//   - Pool: Manages a pool of reusable objects with capacity limits and expiration,
//     useful for connection pools and buffer pools.
//
// Basic usage examples:
//
//	// SingleFlight - prevent cache stampede
//	sf := msync.NewSingleFlight()
//	result, err := sf.Do("cache-key", func() (any, error) {
//	    return queryDatabase()
//	})
//
//	// LockedCalls - serialize operations
//	lc := msync.NewLockedCalls()
//	_, err := lc.Do("user-123", func() (any, error) {
//	    return updateUserBalance(amount)
//	})
//
//	// Limit - control concurrency
//	limit := msync.NewLimit(10) // max 10 concurrent
//	limit.Borrow()
//	defer limit.Return()
//	// perform operation
//
//	// Pool - object pool
//	pool := msync.NewPool(50,
//	    func() any { return createConnection() },
//	    func(x any) { x.(*Connection).Close() },
//	    msync.WithMaxAge(5*time.Minute),
//	)
//	conn := pool.Get()
//	defer pool.Put(conn)
//	// use connection
package msync
