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
	wg         sync.WaitGroup
	val        any
	err        error
	panicValue any
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
	c, fresh := sf.start(key)
	if !fresh {
		return c.result()
	}

	return sf.execute(key, c, fn)
}

// DoEx is like Do but returns whether the result is fresh (newly executed).
// The fresh boolean will be true if the caller executed the function,
// or false if it waited for another caller's result.
func (sf *SingleFlight) DoEx(key string, fn func() (any, error)) (val any, fresh bool, err error) {
	c, fresh := sf.start(key)
	if !fresh {
		val, err = c.result()
		return val, false, err
	}

	val, err = sf.execute(key, c, fn)
	return val, true, err
}

// start returns the call associated with key and reports whether the caller
// is responsible for executing it.
func (sf *SingleFlight) start(key string) (*call, bool) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if c, ok := sf.calls[key]; ok {
		return c, false
	}

	c := &call{}
	c.wg.Add(1)
	sf.calls[key] = c
	return c, true
}

// execute runs fn and always releases callers waiting for the same key. A
// panic is recorded before the waiters are released so every caller observes
// the same outcome.
func (sf *SingleFlight) execute(key string, c *call, fn func() (any, error)) (any, error) {
	defer func() {
		if panicValue := recover(); panicValue != nil {
			c.panicValue = panicValue
		}

		sf.mu.Lock()
		delete(sf.calls, key)
		sf.mu.Unlock()
		c.wg.Done()

		if c.panicValue != nil {
			panic(c.panicValue)
		}
	}()

	c.val, c.err = fn()
	return c.val, c.err
}

// result waits for the shared call and reproduces its return or panic value.
func (c *call) result() (any, error) {
	c.wg.Wait()
	if c.panicValue != nil {
		panic(c.panicValue)
	}
	return c.val, c.err
}
