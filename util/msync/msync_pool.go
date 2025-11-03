package msync

import (
	"sync"
	"time"
)

// Pool is an object pool with capacity limit and expiration support.
// Unlike sync.Pool, it:
// - Supports capacity limits
// - Supports object expiration based on idle time
// - Supports custom create and destroy callbacks
// - Will not be cleared by GC
type Pool struct {
	limit   int           // Maximum number of objects
	created int           // Number of created objects
	maxAge  time.Duration // Maximum idle time before object expires
	lock    sync.Mutex    // Protects the pool
	cond    *sync.Cond    // Condition variable for blocking when pool is full
	head    *node         // Head of the linked list of idle objects
	create  func() any    // Function to create new objects
	destroy func(any)     // Function to destroy objects
}

// node represents a node in the idle object linked list.
type node struct {
	item     any
	next     *node
	lastUsed time.Time
}

// PoolOption is a function type for configuring Pool.
type PoolOption func(*Pool)

// WithMaxAge sets the maximum idle time for objects in the pool.
// Objects idle longer than this duration will be destroyed when retrieved.
func WithMaxAge(d time.Duration) PoolOption {
	return func(p *Pool) {
		p.maxAge = d
	}
}

// NewPool creates and returns a new Pool instance.
//
// Parameters:
//   - limit: Maximum number of objects that can exist
//   - create: Function to create new objects
//   - destroy: Function to destroy objects (can be nil)
//   - opts: Optional configuration options
func NewPool(limit int, create func() any, destroy func(any), opts ...PoolOption) *Pool {
	if create == nil {
		panic("msync: Pool create function cannot be nil")
	}
	if destroy == nil {
		destroy = func(any) {} // No-op destroy function
	}

	p := &Pool{
		limit:   limit,
		create:  create,
		destroy: destroy,
	}
	p.cond = sync.NewCond(&p.lock)

	// Apply options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Get retrieves an object from the pool.
// If the pool is empty and the limit hasn't been reached, a new object is created.
// If the pool is empty and the limit has been reached, Get blocks until an object is available.
func (p *Pool) Get() any {
	p.lock.Lock()
	defer p.lock.Unlock()

	for {
		// Case 1: Try to get an idle object
		if p.head != nil {
			head := p.head
			p.head = head.next

			// Check if the object has expired
			if p.maxAge > 0 && time.Since(head.lastUsed) > p.maxAge {
				p.created--
				p.lock.Unlock()
				p.destroy(head.item)
				p.lock.Lock()
				continue // Try to get next object
			}

			return head.item
		}

		// Case 2: Create a new object if under limit
		if p.created < p.limit {
			p.created++
			p.lock.Unlock()
			item := p.create()
			p.lock.Lock()
			return item
		}

		// Case 3: Pool is full, wait for an object to be returned
		p.cond.Wait()
	}
}

// Put returns an object to the pool.
// If x is nil, it is ignored.
func (p *Pool) Put(x any) {
	if x == nil {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	// Add the object to the head of the linked list
	p.head = &node{
		item:     x,
		next:     p.head,
		lastUsed: time.Now(),
	}

	// Wake up one waiting goroutine
	p.cond.Signal()
}

// Size returns the current number of objects in the pool (both idle and in use).
func (p *Pool) Size() int {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.created
}

// Available returns the number of idle objects currently in the pool.
func (p *Pool) Available() int {
	p.lock.Lock()
	defer p.lock.Unlock()

	count := 0
	for n := p.head; n != nil; n = n.next {
		count++
	}
	return count
}

// Clear removes all idle objects from the pool and destroys them.
func (p *Pool) Clear() {
	p.lock.Lock()
	defer p.lock.Unlock()

	for p.head != nil {
		head := p.head
		p.head = head.next
		p.created--
		p.lock.Unlock()
		p.destroy(head.item)
		p.lock.Lock()
	}
}
