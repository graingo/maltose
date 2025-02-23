package minstance

import "sync"

// Container 单例容器
type Container struct {
	instances map[string]any
	mu        sync.RWMutex
}

func New() *Container {
	return &Container{
		instances: make(map[string]any),
	}
}

// Get 获取已存在的实例
func (c *Container) Get(name string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.instances[name]
}

// Set 设置实例
func (c *Container) Set(name string, instance any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instances[name] = instance
}

// GetOrSetFunc 获取实例,如果不存在则通过函数创建
func (c *Container) GetOrSetFunc(name string, fn func() any) any {
	// 先尝试获取
	c.mu.RLock()
	if instance, ok := c.instances[name]; ok {
		c.mu.RUnlock()
		return instance
	}
	c.mu.RUnlock()

	// 加写锁创建
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if instance, ok := c.instances[name]; ok {
		return instance
	}

	// 创建新实例
	instance := fn()
	c.instances[name] = instance
	return instance
}

// Remove 移除实例
func (c *Container) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.instances, name)
}
