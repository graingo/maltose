package mcache

import (
	"context"
	"sync"
	"time"

	"github.com/graingo/maltose/container/mvar"
)

const (
	// Default cleanup interval for expired items.
	cleanupInterval = time.Minute
)

// AdapterMemory is an adapter for memory cache.
// It is thread-safe and supports LRU elimination.
type AdapterMemory struct {
	mu     sync.RWMutex
	data   *memoryData
	lru    *memoryLru
	closed chan struct{}
}

// NewAdapterMemory creates and returns a new memory adapter.
func NewAdapterMemory(capacity ...int) Adapter {
	cap := 0
	if len(capacity) > 0 {
		cap = capacity[0]
	}
	c := &AdapterMemory{
		data:   newMemoryData(),
		lru:    newMemoryLru(cap),
		closed: make(chan struct{}),
	}
	// Start a background goroutine for cleanup
	go c.cleanupLoop()
	return c
}

// cleanupLoop periodically deletes expired items from the cache.
func (c *AdapterMemory) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.clearExpired()
		case <-c.closed:
			return
		}
	}
}

// clearExpired removes expired items from the cache.
func (c *AdapterMemory) clearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, item := range c.data.Data() {
		if !item.e.IsZero() && now.After(item.e) {
			c.remove(key, item)
		}
	}
}

// remove is an internal method that removes an item from the cache.
// It is not thread-safe and must be called within a lock.
func (c *AdapterMemory) remove(key string, item *memoryDataItem) {
	c.data.Remove(key)
	if item.elem != nil {
		c.lru.Remove(item.elem)
	}
}

// evict removes the least recently used item if the cache is full.
// It is not thread-safe and must be called within a lock.
func (c *AdapterMemory) evict() {
	if c.lru.IsFull() {
		if key, ok := c.lru.Pop(); ok {
			c.data.Remove(key)
		}
	}
}

// Close closes the cache and stops the cleanup goroutine.
func (c *AdapterMemory) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	select {
	case <-c.closed:
	// already closed
	default:
		close(c.closed)
	}
	return nil
}

// Set sets cache with `key`-`value` pair.
func (c *AdapterMemory) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value, duration)
	return nil
}

// set is the internal implementation of Set.
// It is not thread-safe and must be called within a lock.
func (c *AdapterMemory) set(key string, value interface{}, duration time.Duration) {
	var expire time.Time
	if duration > 0 {
		expire = time.Now().Add(duration)
	}
	if item := c.data.Get(key); item != nil {
		item.v = value
		item.e = expire
		if item.elem != nil {
			c.lru.Push(item.elem)
		}
	} else {
		c.evict()
		elem := c.lru.NewElement(key)
		c.data.Set(key, &memoryDataItem{v: value, e: expire, elem: elem})
	}
}

// SetMap batch sets cache with key-value pairs by `data` map.
func (c *AdapterMemory) SetMap(ctx context.Context, data map[string]interface{}, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, value := range data {
		c.set(key, value, duration)
	}
	return nil
}

// SetIfNotExist sets cache with `key`-`value` pair if `key` does not exist in the cache.
func (c *AdapterMemory) SetIfNotExist(ctx context.Context, key string, value interface{}, duration time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item := c.data.Get(key); item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			return false, nil
		}
	}
	c.set(key, value, duration)
	return true, nil
}

// SetIfNotExistFunc sets `key` with result of function `f`.
func (c *AdapterMemory) SetIfNotExistFunc(ctx context.Context, key string, f Func, duration time.Duration) (bool, error) {
	c.mu.Lock()
	if item := c.data.Get(key); item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			c.mu.Unlock()
			return false, nil
		}
		c.remove(key, item)
	}
	c.mu.Unlock()

	value, err := f(ctx)
	if err != nil {
		return false, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// Double check
	if item := c.data.Get(key); item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			return false, nil
		}
		c.remove(key, item)
	}

	var expire time.Time
	if duration > 0 {
		expire = time.Now().Add(duration)
	}
	elem := c.lru.NewElement(key)
	c.data.Set(key, &memoryDataItem{v: value, e: expire, elem: elem})
	c.evict()
	return true, nil
}

// SetIfNotExistFuncLock sets `key` with result of function `f` with lock.
func (c *AdapterMemory) SetIfNotExistFuncLock(ctx context.Context, key string, f Func, duration time.Duration) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item := c.data.Get(key); item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			return false, nil
		}
		c.remove(key, item)
	}

	value, err := f(ctx)
	if err != nil {
		return false, err
	}

	var expire time.Time
	if duration > 0 {
		expire = time.Now().Add(duration)
	}
	elem := c.lru.NewElement(key)
	c.data.Set(key, &memoryDataItem{v: value, e: expire, elem: elem})
	c.evict()
	return true, nil
}

// Get retrieves and returns the associated value of given `key`.
func (c *AdapterMemory) Get(ctx context.Context, key string) (*mvar.Var, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := c.data.Get(key)
	if item == nil {
		return nil, nil
	}
	if !item.e.IsZero() && time.Now().After(item.e) {
		c.remove(key, item)
		return nil, nil
	}
	c.lru.Push(item.elem)
	return mvar.New(item.v), nil
}

// GetOrSet retrieves and returns the value of `key`, or sets `key`-`value` pair and returns `value`.
func (c *AdapterMemory) GetOrSet(ctx context.Context, key string, value interface{}, duration time.Duration) (*mvar.Var, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := c.data.Get(key)
	if item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			c.lru.Push(item.elem)
			return mvar.New(item.v), nil
		}
		c.remove(key, item)
	}

	var expire time.Time
	if duration > 0 {
		expire = time.Now().Add(duration)
	}
	elem := c.lru.NewElement(key)
	c.data.Set(key, &memoryDataItem{v: value, e: expire, elem: elem})
	c.evict()
	return mvar.New(value), nil
}

// GetOrSetFunc retrieves and returns the value of `key`, or sets `key` with result of function `f`.
func (c *AdapterMemory) GetOrSetFunc(ctx context.Context, key string, f Func, duration time.Duration) (*mvar.Var, error) {
	// Get without lock
	if v, _ := c.Get(ctx, key); v != nil {
		return v, nil
	}

	// Lock and double check
	c.mu.Lock()
	if item := c.data.Get(key); item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			c.lru.Push(item.elem)
			c.mu.Unlock()
			return mvar.New(item.v), nil
		}
		c.remove(key, item)
	}
	c.mu.Unlock()

	value, err := f(ctx)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// Double check again
	if item := c.data.Get(key); item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			c.lru.Push(item.elem)
			return mvar.New(item.v), nil
		}
		c.remove(key, item)
	}

	var expire time.Time
	if duration > 0 {
		expire = time.Now().Add(duration)
	}
	elem := c.lru.NewElement(key)
	c.data.Set(key, &memoryDataItem{v: value, e: expire, elem: elem})
	c.evict()
	return mvar.New(value), nil
}

// GetOrSetFuncLock retrieves and returns the value of `key`, or sets `key` with result of function `f` with lock.
func (c *AdapterMemory) GetOrSetFuncLock(ctx context.Context, key string, f Func, duration time.Duration) (*mvar.Var, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item := c.data.Get(key)
	if item != nil {
		if item.e.IsZero() || time.Now().Before(item.e) {
			c.lru.Push(item.elem)
			return mvar.New(item.v), nil
		}
		c.remove(key, item)
	}

	value, err := f(ctx)
	if err != nil {
		return nil, err
	}

	var expire time.Time
	if duration > 0 {
		expire = time.Now().Add(duration)
	}
	elem := c.lru.NewElement(key)
	c.data.Set(key, &memoryDataItem{v: value, e: expire, elem: elem})
	c.evict()
	return mvar.New(value), nil
}

// Contains checks and returns true if `key` exists in the cache, or else returns false.
func (c *AdapterMemory) Contains(ctx context.Context, key string) (bool, error) {
	v, err := c.Get(ctx, key)
	return v != nil, err
}

// Remove deletes one or more keys from cache.
func (c *AdapterMemory) Remove(ctx context.Context, keys ...string) (*mvar.Var, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var lastValue *mvar.Var
	for _, key := range keys {
		if item := c.data.Get(key); item != nil {
			lastValue = mvar.New(item.v)
			c.remove(key, item)
		}
	}
	return lastValue, nil
}

// Data returns a copy of all key-value pairs in the cache as map type.
func (c *AdapterMemory) Data(ctx context.Context) (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data := make(map[string]interface{})
	for k, v := range c.data.Data() {
		if v.e.IsZero() || time.Now().Before(v.e) {
			data[k] = v.v
		}
	}
	return data, nil
}

// Keys returns all keys in the cache as slice.
func (c *AdapterMemory) Keys(ctx context.Context) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]string, 0, c.lru.Len())
	for elem := c.lru.list.Front(); elem != nil; elem = elem.Next() {
		keys = append(keys, elem.Value.(string))
	}
	return keys, nil
}

// Values returns all values in the cache as slice.
func (c *AdapterMemory) Values(ctx context.Context) ([]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := make([]interface{}, 0, c.lru.Len())
	for elem := c.lru.list.Front(); elem != nil; elem = elem.Next() {
		key := elem.Value.(string)
		if item := c.data.Get(key); item != nil {
			if item.e.IsZero() || time.Now().Before(item.e) {
				values = append(values, item.v)
			}
		}
	}
	return values, nil
}

// Size returns the size of the cache.
func (c *AdapterMemory) Size(ctx context.Context) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lru.Len(), nil
}

// Update updates the value of `key` without changing its expiration and returns the old value.
func (c *AdapterMemory) Update(ctx context.Context, key string, value interface{}) (oldValue *mvar.Var, exist bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item := c.data.Get(key); item != nil {
		if !item.e.IsZero() && time.Now().After(item.e) {
			c.remove(key, item)
			return nil, false, nil
		}
		oldValue = mvar.New(item.v)
		item.v = value
		c.lru.Push(item.elem)
		return oldValue, true, nil
	}
	return nil, false, nil
}

// UpdateExpire updates the expiration of `key` and returns the old expiration duration value.
func (c *AdapterMemory) UpdateExpire(ctx context.Context, key string, duration time.Duration) (oldDuration time.Duration, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item := c.data.Get(key); item != nil {
		if !item.e.IsZero() && time.Now().After(item.e) {
			c.remove(key, item)
			return -1, nil
		}
		if !item.e.IsZero() {
			oldDuration = time.Until(item.e)
		}
		item.e = time.Now().Add(duration)
		c.lru.Push(item.elem)
		return oldDuration, nil
	}
	return -1, nil
}

// GetExpire retrieves and returns the expiration of `key` in the cache.
func (c *AdapterMemory) GetExpire(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if item := c.data.Get(key); item != nil {
		if !item.e.IsZero() && time.Now().After(item.e) {
			return -1, nil
		}
		if item.e.IsZero() {
			// Never expires. It is represented as 0 in Go.
			return 0, nil
		}
		return time.Until(item.e), nil
	}
	return -1, nil
}

// Clear clears all data of the cache.
func (c *AdapterMemory) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Clear()
	c.data.Clear()
	return nil
}
