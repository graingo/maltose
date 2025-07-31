package mcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mcache"
	"github.com/stretchr/testify/assert"
)

// TestAdapterMemory_BasicOperations covers basic Set, Get, and expiration.
func TestAdapterMemory_BasicOperations(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	// Test Set and Get
	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "value1", v.String())

	// Test expiration
	time.Sleep(150 * time.Millisecond)
	v, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// Test Get non-existent key
	v, err = cache.Get(ctx, "non-existent-key")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

// TestAdapterMemory_MapOperations covers batch operations like SetMap, Data, Keys, Values.
func TestAdapterMemory_MapOperations(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	// Test SetMap
	data := map[string]interface{}{"key1": "value1", "key2": 2}
	err := cache.SetMap(ctx, data, 0)
	assert.NoError(t, err)

	// Test Data
	retrievedData, err := cache.Data(ctx)
	assert.NoError(t, err)
	assert.Equal(t, data, retrievedData)

	// Test Keys
	keys, err := cache.Keys(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"key1", "key2"}, keys)

	// Test Values
	values, err := cache.Values(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []interface{}{"value1", 2}, values)
}

// TestAdapterMemory_AtomicSetOperations covers SetIfNotExist and its variants.
func TestAdapterMemory_AtomicSetOperations(t *testing.T) {
	t.Run("SetIfNotExist", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()

		// Set for the first time
		ok, err := cache.SetIfNotExist(ctx, "key1", "value1", 50*time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, ok)

		// Try to set again, should fail
		ok, err = cache.SetIfNotExist(ctx, "key1", "value2", 0)
		assert.NoError(t, err)
		assert.False(t, ok)
		v, _ := cache.Get(ctx, "key1")
		assert.Equal(t, "value1", v.String())

		// Set again after expiration
		time.Sleep(100 * time.Millisecond)
		ok, err = cache.SetIfNotExist(ctx, "key1", "value2", 0)
		assert.NoError(t, err)
		assert.True(t, ok)
		v, _ = cache.Get(ctx, "key1")
		assert.Equal(t, "value2", v.String())
	})

	t.Run("SetIfNotExistFunc", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		called := false
		f := func(_ context.Context) (interface{}, error) {
			called = true
			return "value1", nil
		}

		ok, err := cache.SetIfNotExistFunc(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, called)

		called = false // reset flag
		ok, err = cache.SetIfNotExistFunc(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.False(t, called)
	})

	t.Run("SetIfNotExistFuncLock", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		called := false
		f := func(_ context.Context) (interface{}, error) {
			called = true
			return "value1", nil
		}
		ok, err := cache.SetIfNotExistFuncLock(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.True(t, called)

		called = false
		ok, err = cache.SetIfNotExistFuncLock(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.False(t, called)
	})

	t.Run("SetIfNotExistFunc_Error", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		f := func(_ context.Context) (interface{}, error) {
			return nil, assert.AnError
		}
		ok, err := cache.SetIfNotExistFunc(ctx, "key1", f, 0)
		assert.Error(t, err)
		assert.False(t, ok)
	})
}

// TestAdapterMemory_AtomicGetOperations covers GetOrSet and its variants.
func TestAdapterMemory_AtomicGetOperations(t *testing.T) {
	t.Run("GetOrSet", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()

		v, err := cache.GetOrSet(ctx, "key1", "value1", 0)
		assert.NoError(t, err)
		assert.Equal(t, "value1", v.String())

		v, err = cache.GetOrSet(ctx, "key1", "value2", 0)
		assert.NoError(t, err)
		assert.Equal(t, "value1", v.String())
	})

	t.Run("GetOrSetFunc", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		called := false
		f := func(_ context.Context) (interface{}, error) {
			called = true
			return "value1", nil
		}

		v, err := cache.GetOrSetFunc(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "value1", v.String())

		called = false // reset flag
		v, err = cache.GetOrSetFunc(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.False(t, called)
		assert.Equal(t, "value1", v.String())
	})

	t.Run("GetOrSetFuncLock", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		called := false
		f := func(_ context.Context) (interface{}, error) {
			called = true
			return "value1", nil
		}
		v, err := cache.GetOrSetFuncLock(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "value1", v.String())

		called = false
		v, err = cache.GetOrSetFuncLock(ctx, "key1", f, 0)
		assert.NoError(t, err)
		assert.False(t, called)
		assert.Equal(t, "value1", v.String())
	})

	t.Run("GetOrSetFunc_Error", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		f := func(_ context.Context) (interface{}, error) {
			return nil, assert.AnError
		}
		v, err := cache.GetOrSetFunc(ctx, "key1", f, 0)
		assert.Error(t, err)
		assert.Nil(t, v)
	})
}

// TestAdapterMemory_MetadataOperations covers operations that inspect or modify cache metadata.
func TestAdapterMemory_MetadataOperations(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	_ = cache.Set(ctx, "key2", "value2", 0)

	// Test Contains
	ok, err := cache.Contains(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, ok)
	ok, err = cache.Contains(ctx, "non-existent")
	assert.NoError(t, err)
	assert.False(t, ok)

	// Test Size
	size, err := cache.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, size)

	// Test GetExpire
	d, err := cache.GetExpire(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, d > 0 && d <= 100*time.Millisecond)
	d, err = cache.GetExpire(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), d)
}

// TestAdapterMemory_UpdateOperations covers Update and UpdateExpire.
func TestAdapterMemory_UpdateOperations(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", 100*time.Millisecond)

	// Test Update
	oldVal, exist, err := cache.Update(ctx, "key1", "newValue")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "value1", oldVal.String())
	newVal, _ := cache.Get(ctx, "key1")
	assert.Equal(t, "newValue", newVal.String())

	// Test expiration is not changed by Update
	time.Sleep(150 * time.Millisecond)
	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// Test UpdateExpire
	_ = cache.Set(ctx, "key2", "value2", 100*time.Millisecond)
	oldDur, err := cache.UpdateExpire(ctx, "key2", 200*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, oldDur > 0 && oldDur <= 100*time.Millisecond)
	time.Sleep(150 * time.Millisecond)
	v, err = cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.NotNil(t, v) // Should still exist
}

// TestAdapterMemory_DeleteOperations covers Remove and Clear.
func TestAdapterMemory_DeleteOperations(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", 2, 0)
	_ = cache.Set(ctx, "key3", "value3", 0)

	// Test Remove single key
	lastVal, err := cache.Remove(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", lastVal.String())
	size, _ := cache.Size(ctx)
	assert.Equal(t, 2, size)

	// Test Remove multiple keys
	lastVal, err = cache.Remove(ctx, "key2", "key3")
	assert.NoError(t, err)
	assert.Contains(t, []interface{}{2, "value3"}, lastVal.Val())
	size, _ = cache.Size(ctx)
	assert.Equal(t, 0, size)

	// Test Clear
	_ = cache.Set(ctx, "key1", "value1", 0)
	err = cache.Clear(ctx)
	assert.NoError(t, err)
	size, _ = cache.Size(ctx)
	assert.Equal(t, 0, size)
}

// TestAdapterMemory_LRU verifies the Least Recently Used eviction policy.
func TestAdapterMemory_LRU(t *testing.T) {
	cache := mcache.NewAdapterMemory(2) // capacity is 2
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)

	// Access key1 to make it recently used
	_, _ = cache.Get(ctx, "key1")

	_ = cache.Set(ctx, "key3", "value3", 0)

	// key2 should be evicted because it was the least recently used.
	v1, _ := cache.Get(ctx, "key1")
	assert.NotNil(t, v1)
	v2, _ := cache.Get(ctx, "key2")
	assert.Nil(t, v2)
	v3, _ := cache.Get(ctx, "key3")
	assert.NotNil(t, v3)

	// Add another key, key1 should now be the LRU and get evicted.
	_ = cache.Set(ctx, "key4", "value4", 0)
	v1, _ = cache.Get(ctx, "key1")
	assert.Nil(t, v1)
}

// TestAdapterMemory_EdgeCases covers various edge cases.
func TestAdapterMemory_EdgeCases(t *testing.T) {
	t.Run("Close", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		err := cache.Close(ctx)
		assert.NoError(t, err)
		// can be closed multiple times
		err = cache.Close(ctx)
		assert.NoError(t, err)
	})

	t.Run("Remove_NonExistent", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		lastValue, err := cache.Remove(ctx, "non-existent-key")
		assert.NoError(t, err)
		assert.Nil(t, lastValue)
	})

	t.Run("Update_NonExistent", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		oldValue, exist, err := cache.Update(ctx, "non-existent-key", "value")
		assert.NoError(t, err)
		assert.False(t, exist)
		assert.Nil(t, oldValue)
	})

	t.Run("UpdateExpire_NonExistent", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		oldDuration, err := cache.UpdateExpire(ctx, "non-existent-key", time.Second)
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-1), oldDuration)
	})

	t.Run("GetExpire_NonExistent", func(t *testing.T) {
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()
		duration, err := cache.GetExpire(ctx, "non-existent-key")
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-1), duration)
	})

	t.Run("Set_NilValue", func(t *testing.T) {
		// In current implementation, setting nil just sets the value to nil.
		cache := mcache.NewAdapterMemory()
		ctx := context.Background()

		_ = cache.Set(ctx, "key1", "value1", 0)
		_ = cache.Set(ctx, "key1", nil, 0)
		v, err := cache.Get(ctx, "key1")
		assert.NoError(t, err)
		assert.NotNil(t, v)
		assert.True(t, v.IsNil())
		assert.Nil(t, v.Val())
	})
}
