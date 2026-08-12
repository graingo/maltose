package msync

import "errors"

var (
	// ErrLimitReturn is returned when Return is called without a corresponding Borrow.
	ErrLimitReturn = errors.New("msync: limit return without borrow")
)

// Limit is a semaphore implementation using channels to limit concurrent execution.
// It allows controlling the maximum number of concurrent operations.
type Limit struct {
	pool chan struct{}
}

// NewLimit creates and returns a new Limit instance with the specified capacity.
// The capacity determines the maximum number of concurrent operations allowed.
func NewLimit(n int) *Limit {
	return &Limit{
		pool: make(chan struct{}, n),
	}
}

// Borrow acquires a slot from the limit pool, blocking if the pool is full.
// It must be paired with a Return call to release the slot.
func (l *Limit) Borrow() {
	l.pool <- struct{}{}
}

// TryBorrow attempts to acquire a slot from the limit pool without blocking.
// It returns true if successful, false if the pool is full.
func (l *Limit) TryBorrow() bool {
	select {
	case l.pool <- struct{}{}:
		return true
	default:
		return false
	}
}

// Return releases a slot back to the limit pool.
// It returns an error if Return is called more times than Borrow.
func (l *Limit) Return() error {
	select {
	case <-l.pool:
		return nil
	default:
		return ErrLimitReturn
	}
}
