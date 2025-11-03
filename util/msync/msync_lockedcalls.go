package msync

import "sync"

// LockedCalls ensures that calls with the same key are executed sequentially.
// Unlike SingleFlight, each call executes the function independently and gets its own result.
// This is useful for write operations where each operation must be executed,
// but operations with the same key must be serialized.
type LockedCalls struct {
	mu sync.Mutex
	m  map[string]*sync.WaitGroup
}

// NewLockedCalls creates and returns a new LockedCalls instance.
func NewLockedCalls() *LockedCalls {
	return &LockedCalls{
		m: make(map[string]*sync.WaitGroup),
	}
}

// Do executes the given function for the specified key.
// If another goroutine is already executing a function for the same key,
// this call will wait for it to complete before executing.
// Each call executes the function independently and receives its own result.
func (lc *LockedCalls) Do(key string, fn func() (any, error)) (any, error) {
begin:
	lc.mu.Lock()

	// Check if another goroutine is processing this key
	if wg, ok := lc.m[key]; ok {
		lc.mu.Unlock()
		wg.Wait()  // Wait for it to complete
		goto begin // Try again to acquire the lock
	}

	// This goroutine gets to process the key
	return lc.makeCall(key, fn)
}

// makeCall executes the function and manages the lock lifecycle.
func (lc *LockedCalls) makeCall(key string, fn func() (any, error)) (any, error) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	lc.m[key] = wg
	lc.mu.Unlock()

	// Execute the function
	val, err := fn()

	// Clean up: Remove key from map first, then signal completion
	// This order is important to avoid race conditions
	lc.mu.Lock()
	delete(lc.m, key)
	lc.mu.Unlock()
	wg.Done()

	return val, err
}
