package mcache

// memoryData is the underlying data structure for the memory cache.
// Note that this structure is not thread-safe.
type memoryData struct {
	data map[string]*memoryDataItem
}

// newMemoryData creates and returns a new memoryData.
func newMemoryData() *memoryData {
	return &memoryData{
		data: make(map[string]*memoryDataItem),
	}
}

// Set sets a key-value pair.
func (md *memoryData) Set(key string, item *memoryDataItem) {
	md.data[key] = item
}

// Get retrieves an item by key. It returns nil if the key does not exist.
func (md *memoryData) Get(key string) *memoryDataItem {
	return md.data[key]
}

// Remove deletes a key-value pair.
func (md *memoryData) Remove(key string) (item *memoryDataItem) {
	if item, ok := md.data[key]; ok {
		delete(md.data, key)
		return item
	}
	return nil
}

// Data returns a copy of all key-value pairs.
func (md *memoryData) Data() map[string]*memoryDataItem {
	m := make(map[string]*memoryDataItem, len(md.data))
	for k, v := range md.data {
		m[k] = v
	}
	return m
}

// Clear removes all items from the map.
func (md *memoryData) Clear() {
	md.data = make(map[string]*memoryDataItem)
}
