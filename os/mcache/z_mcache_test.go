package mcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mcache"
	"github.com/stretchr/testify/assert"
)

func TestCache_New(t *testing.T) {
	cache := mcache.New()
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.GetAdapter())
}

type mockAdapterForTest struct {
	mcache.Adapter
}

func TestCache_NewWithAdapter(t *testing.T) {
	adapter := &mockAdapterForTest{}
	cache := mcache.NewWithAdapter(adapter)
	assert.NotNil(t, cache)
	assert.Equal(t, adapter, cache.GetAdapter())
}

func TestCache_SetAdapter(t *testing.T) {
	cache := mcache.New()
	adapter := &mockAdapterForTest{}
	cache.SetAdapter(adapter)
	assert.Equal(t, adapter, cache.GetAdapter())
}

func TestCache_Removes(t *testing.T) {
	ctx := context.Background()
	cache := mcache.New()
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)

	err := cache.Removes(ctx, []string{"key1", "key2"})
	assert.NoError(t, err)

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestCache_KeyStrings(t *testing.T) {
	ctx := context.Background()
	cache := mcache.New()
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)

	keys, err := cache.KeyStrings(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"key1", "key2"}, keys)
}

// Test default cache functions from mcache.go
func TestDefaultCache(t *testing.T) {
	ctx := context.Background()

	// Clear at start to ensure clean state
	err := mcache.Clear(ctx)
	assert.NoError(t, err)

	// Set and Get
	err = mcache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)
	v, err := mcache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "value1", v.String())

	// Contains
	ok, err := mcache.Contains(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, ok)

	// Size
	size, err := mcache.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, size)

	// Data
	data, err := mcache.Data(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"key1": "value1"}, data)

	// Keys
	keys, err := mcache.Keys(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"key1"}, keys)

	// Values
	values, err := mcache.Values(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []interface{}{"value1"}, values)

	// Update
	oldVal, exist, err := mcache.Update(ctx, "key1", "newValue")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "value1", oldVal.String())

	// GetExpire
	d, err := mcache.GetExpire(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, d > 0)

	// UpdateExpire
	_, err = mcache.UpdateExpire(ctx, "key1", 200*time.Millisecond)
	assert.NoError(t, err)

	// Remove
	lastVal, err := mcache.Remove(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "newValue", lastVal.String())
	size, err = mcache.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)

	// GetOrSet
	v, err = mcache.GetOrSet(ctx, "key2", "value2", 0)
	assert.NoError(t, err)
	assert.Equal(t, "value2", v.String())

	// Clear
	err = mcache.Clear(ctx)
	assert.NoError(t, err)
	size, err = mcache.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
}

func TestDefaultCache_SetMap(t *testing.T) {
	ctx := context.Background()
	err := mcache.Clear(ctx)
	assert.NoError(t, err)

	data := map[string]interface{}{"a": 1, "b": "c"}
	err = mcache.SetMap(ctx, data, 0)
	assert.NoError(t, err)

	size, err := mcache.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, size)

	dataFromCache, err := mcache.Data(ctx)
	assert.NoError(t, err)
	assert.Equal(t, data, dataFromCache)
}

func TestDefaultCache_SetIfNotExist(t *testing.T) {
	ctx := context.Background()
	err := mcache.Clear(ctx)
	assert.NoError(t, err)

	ok, err := mcache.SetIfNotExist(ctx, "k", "v", 0)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = mcache.SetIfNotExist(ctx, "k", "v", 0)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestDefaultCache_SetIfNotExistFunc(t *testing.T) {
	ctx := context.Background()
	err := mcache.Clear(ctx)
	assert.NoError(t, err)
	called := false
	f := func(ctx context.Context) (interface{}, error) {
		called = true
		return "v", nil
	}
	ok, err := mcache.SetIfNotExistFunc(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, called)

	called = false
	ok, err = mcache.SetIfNotExistFunc(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, called)
}

func TestDefaultCache_SetIfNotExistFuncLock(t *testing.T) {
	ctx := context.Background()
	err := mcache.Clear(ctx)
	assert.NoError(t, err)
	called := false
	f := func(ctx context.Context) (interface{}, error) {
		called = true
		return "v", nil
	}
	ok, err := mcache.SetIfNotExistFuncLock(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, called)

	called = false
	ok, err = mcache.SetIfNotExistFuncLock(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, called)
}

func TestDefaultCache_GetOrSetFunc(t *testing.T) {
	ctx := context.Background()
	err := mcache.Clear(ctx)
	assert.NoError(t, err)

	called := false
	f := func(ctx context.Context) (interface{}, error) {
		called = true
		return "v", nil
	}

	v, err := mcache.GetOrSetFunc(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "v", v.String())

	called = false
	v, err = mcache.GetOrSetFunc(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, "v", v.String())
}

func TestDefaultCache_GetOrSetFuncLock(t *testing.T) {
	ctx := context.Background()
	err := mcache.Clear(ctx)
	assert.NoError(t, err)

	called := false
	f := func(ctx context.Context) (interface{}, error) {
		called = true
		return "v", nil
	}

	v, err := mcache.GetOrSetFuncLock(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "v", v.String())

	called = false
	v, err = mcache.GetOrSetFuncLock(ctx, "k", f, 0)
	assert.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, "v", v.String())
}
