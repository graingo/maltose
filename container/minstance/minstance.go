package minstance

import "sync"

// Container is a singleton container.
type Container struct {
	instances map[string]any
	mu        sync.RWMutex
}

func New() *Container {
	return &Container{
		instances: make(map[string]any),
	}
}

// Get gets the existing instance.
func (c *Container) Get(name string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.instances[name]
}

// Set sets the instance.
func (c *Container) Set(name string, instance any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instances[name] = instance
}

// GetOrSetFunc gets the instance, if it does not exist, it will be created by the function.
func (c *Container) GetOrSetFunc(name string, fn func() any) any {
	// try to get the instance
	c.mu.RLock()
	if instance, ok := c.instances[name]; ok {
		c.mu.RUnlock()
		return instance
	}
	c.mu.RUnlock()

	// add write lock to create
	c.mu.Lock()
	defer c.mu.Unlock()

	// double check
	if instance, ok := c.instances[name]; ok {
		return instance
	}

	// create new instance
	instance := fn()
	c.instances[name] = instance
	return instance
}

// Remove removes the instance.
func (c *Container) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.instances, name)
}

// All returns all instances.
func (c *Container) All() []any {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list := make([]any, 0, len(c.instances))
	for _, v := range c.instances {
		list = append(list, v)
	}
	return list
}

// Count returns the number of instances.
func (c *Container) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.instances)
}
