package mcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mcache"
	"github.com/stretchr/testify/assert"
)

// --- Initialization and Configuration ---

type mockAdapterForTest struct {
	mcache.Adapter
}

func TestCache_Initialization(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		cache := mcache.New()
		assert.NotNil(t, cache)
		assert.NotNil(t, cache.GetAdapter())
	})

	t.Run("new_with_adapter", func(t *testing.T) {
		adapter := &mockAdapterForTest{}
		cache := mcache.NewWithAdapter(adapter)
		assert.NotNil(t, cache)
		assert.Equal(t, adapter, cache.GetAdapter())
	})

	t.Run("set_adapter", func(t *testing.T) {
		cache := mcache.New()
		adapter := &mockAdapterForTest{}
		cache.SetAdapter(adapter)
		assert.Equal(t, adapter, cache.GetAdapter())
	})
}

// --- Default Cache Operations ---

func TestDefaultCacheOperations(t *testing.T) {
	ctx := context.Background()

	// Helper to ensure clean state for each sub-test.
	setup := func(t *testing.T) {
		t.Helper()
		err := mcache.Clear(ctx)
		assert.NoError(t, err)
	}

	t.Run("basic_crud", func(t *testing.T) {
		setup(t)
		// Set and Get
		err := mcache.Set(ctx, "key1", "value1", 100*time.Millisecond)
		assert.NoError(t, err)
		v, err := mcache.Get(ctx, "key1")
		assert.NoError(t, err)
		assert.NotNil(t, v)
		assert.Equal(t, "value1", v.String())

		// Contains
		ok, err := mcache.Contains(ctx, "key1")
		assert.NoError(t, err)
		assert.True(t, ok)

		// Update
		oldVal, exist, err := mcache.Update(ctx, "key1", "newValue")
		assert.NoError(t, err)
		assert.True(t, exist)
		assert.Equal(t, "value1", oldVal.String())

		// Remove
		lastVal, err := mcache.Remove(ctx, "key1")
		assert.NoError(t, err)
		assert.Equal(t, "newValue", lastVal.String())
	})

	t.Run("map_operations", func(t *testing.T) {
		setup(t)
		data := map[string]interface{}{"a": 1, "b": "c"}
		err := mcache.SetMap(ctx, data, 0)
		assert.NoError(t, err)

		size, err := mcache.Size(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, size)

		dataFromCache, err := mcache.Data(ctx)
		assert.NoError(t, err)
		assert.Equal(t, data, dataFromCache)
	})

	t.Run("set_if_not_exist", func(t *testing.T) {
		setup(t)
		ok, err := mcache.SetIfNotExist(ctx, "k", "v", 0)
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = mcache.SetIfNotExist(ctx, "k", "v", 0)
		assert.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("set_if_not_exist_func", func(t *testing.T) {
		setup(t)
		called := false
		f := func(_ context.Context) (interface{}, error) {
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
	})

	t.Run("set_if_not_exist_func_lock", func(t *testing.T) {
		setup(t)
		called := false
		f := func(_ context.Context) (interface{}, error) {
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
	})

	t.Run("get_or_set_func", func(t *testing.T) {
		setup(t)
		called := false
		f := func(_ context.Context) (interface{}, error) {
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
	})

	t.Run("get_or_set_func_lock", func(t *testing.T) {
		setup(t)
		called := false
		f := func(_ context.Context) (interface{}, error) {
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
	})

	t.Run("metadata_and_bulk_operations", func(t *testing.T) {
		setup(t)
		cache := mcache.New() // Create an instance for instance methods
		_ = cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
		_ = cache.Set(ctx, "key2", "value2", 0)

		// Size
		size, err := cache.Size(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, size)

		// KeyStrings
		keyStrings, err := cache.KeyStrings(ctx)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keyStrings)

		// Keys
		keys, err := cache.Keys(ctx)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"key1", "key2"}, keys)

		// Values
		values, err := cache.Values(ctx)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []interface{}{"value1", "value2"}, values)

		// Removes
		err = cache.Removes(ctx, []string{"key1", "key2"})
		assert.NoError(t, err)
		size, _ = cache.Size(ctx)
		assert.Equal(t, 0, size)
	})

	t.Run("expiration", func(t *testing.T) {
		setup(t)
		_ = mcache.Set(ctx, "key1", "value1", 100*time.Millisecond)

		// GetExpire and UpdateExpire
		d, err := mcache.GetExpire(ctx, "key1")
		assert.NoError(t, err)
		assert.True(t, d > 0)

		_, err = mcache.UpdateExpire(ctx, "key1", 200*time.Millisecond)
		assert.NoError(t, err)

		d2, err := mcache.GetExpire(ctx, "key1")
		assert.NoError(t, err)
		assert.True(t, d2 > d)
	})
}
