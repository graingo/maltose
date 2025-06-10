package mcache

import (
	"container/list"
)

// memoryLru is the LRU manager for memory cache.
// Note that this structure is not thread-safe.
type memoryLru struct {
	list *list.List
	cap  int
}

// newMemoryLru creates and returns a new LRU manager.
func newMemoryLru(capacity int) *memoryLru {
	return &memoryLru{
		list: list.New(),
		cap:  capacity,
	}
}

// Push pushes a new element to the front of the LRU list.
// If the element already exists, it moves it to the front.
func (lru *memoryLru) Push(elem *list.Element) {
	if lru.cap <= 0 {
		return
	}
	lru.list.MoveToFront(elem)
}

// Pop removes and returns the key from the back of the LRU list (least recently used).
func (lru *memoryLru) Pop() (key string, ok bool) {
	if lru.cap <= 0 {
		return
	}
	if elem := lru.list.Back(); elem != nil {
		key, ok = lru.list.Remove(elem).(string)
	}
	return
}

// Remove removes a specific element from the LRU list.
func (lru *memoryLru) Remove(elem *list.Element) {
	if lru.cap <= 0 {
		return
	}
	lru.list.Remove(elem)
}

// Len returns the number of items in the LRU list.
func (lru *memoryLru) Len() int {
	return lru.list.Len()
}

// IsFull checks if the LRU list is full.
func (lru *memoryLru) IsFull() bool {
	if lru.cap <= 0 {
		return false
	}
	return lru.list.Len() >= lru.cap
}

// NewElement creates a new element for the LRU list.
func (lru *memoryLru) NewElement(key string) *list.Element {
	if lru.cap <= 0 {
		return nil
	}
	return lru.list.PushFront(key)
}

// Clear removes all items from the LRU list.
func (lru *memoryLru) Clear() {
	lru.list.Init()
}
