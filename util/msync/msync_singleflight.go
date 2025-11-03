package msync

import "sync"

// SingleFlight prevents duplicate function calls for the same key.
// Multiple concurrent calls with the same key will share the result of a single execution.
type SingleFlight struct {
	mu    sync.Mutex
	calls map[string]*call
}

// call represents an in-flight or completed Do call.
type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// NewSingleFlight creates and returns a new SingleFlight instance.
func NewSingleFlight() *SingleFlight {
	return &SingleFlight{
		calls: make(map[string]*call),
	}
}

// Do executes and returns the results of the given function,
// making sure that only one execution is in-flight for a given key at a time.
// If a duplicate comes in, the duplicate caller waits for the original to complete
// and receives the same results.
func (sf *SingleFlight) Do(key string, fn func() (any, error)) (any, error) {
	sf.mu.Lock()

	// Check if there's already an in-flight call for this key
	if c, ok := sf.calls[key]; ok {
		sf.mu.Unlock()
		c.wg.Wait() // Wait for the in-flight call to complete
		return c.val, c.err
	}

	// First call for this key, create a new call
	c := &call{}
	c.wg.Add(1)
	sf.calls[key] = c
	sf.mu.Unlock()

	// Execute the function
	c.val, c.err = fn()

	// Clean up and notify waiting goroutines
	sf.mu.Lock()
	delete(sf.calls, key)
	sf.mu.Unlock()
	c.wg.Done()

	return c.val, c.err
}

// DoEx is like Do but returns whether the result is fresh (newly executed).
// The fresh boolean will be true if the caller executed the function,
// or false if it waited for another caller's result.
func (sf *SingleFlight) DoEx(key string, fn func() (any, error)) (val any, fresh bool, err error) {
	sf.mu.Lock()

	if c, ok := sf.calls[key]; ok {
		sf.mu.Unlock()
		c.wg.Wait()
		return c.val, false, c.err // Shared result
	}

	c := &call{}
	c.wg.Add(1)
	sf.calls[key] = c
	sf.mu.Unlock()

	c.val, c.err = fn()

	sf.mu.Lock()
	delete(sf.calls, key)
	sf.mu.Unlock()
	c.wg.Done()

	return c.val, true, c.err // Fresh result
}
