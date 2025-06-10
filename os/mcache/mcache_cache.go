package mcache

import (
	"context"
)

// Cache struct.
type Cache struct {
	adapter
}

// adapter is alias of Adapter, for embedded attribute purpose only.
type adapter = Adapter

// New creates and returns a new cache object using default memory adapter.
// Note that the LRU feature is only available using memory adapter.
func New(lruCap ...int) *Cache {
	var capacity int
	if len(lruCap) > 0 {
		capacity = lruCap[0]
	}
	c := &Cache{
		adapter: NewAdapterMemory(capacity),
	}
	return c
}

// NewWithAdapter creates and returns a Cache object with given Adapter implements.
func NewWithAdapter(adapter Adapter) *Cache {
	return &Cache{
		adapter: adapter,
	}
}

// SetAdapter changes the adapter for this cache.
func (c *Cache) SetAdapter(adapter Adapter) {
	c.adapter = adapter
}

// GetAdapter returns the adapter that is set in current Cache.
func (c *Cache) GetAdapter() Adapter {
	return c.adapter
}

// Removes deletes `keys` in the cache.
// It is a wrapper of `Remove` method.
func (c *Cache) Removes(ctx context.Context, keys []string) error {
	_, err := c.Remove(ctx, keys...)
	return err
}

// KeyStrings returns all keys in the cache as string slice.
func (c *Cache) KeyStrings(ctx context.Context) ([]string, error) {
	return c.Keys(ctx)
}
