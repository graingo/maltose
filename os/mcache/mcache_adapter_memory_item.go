package mcache

import (
	"container/list"
	"time"
)

// memoryDataItem is the internal data structure for memory cache.
// It holds the value, expiration time, and a pointer to its corresponding element in the LRU list.
type memoryDataItem struct {
	v    interface{}   // Value.
	e    time.Time     // Expire time.
	elem *list.Element // Pointer to the item in the LRU list.
}
