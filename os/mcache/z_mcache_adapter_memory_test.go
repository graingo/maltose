package mcache_test

import (
	"context"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mcache"
	"github.com/stretchr/testify/assert"
)

func TestAdapterMemory_SetAndGet(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "value1", v.String())

	time.Sleep(150 * time.Millisecond)
	v, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestAdapterMemory_SetMap(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	data := map[string]interface{}{"key1": "value1", "key2": 2}
	err := cache.SetMap(ctx, data, 100*time.Millisecond)
	assert.NoError(t, err)

	v1, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v1)
	assert.Equal(t, "value1", v1.String())

	v2, err := cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.NotNil(t, v2)
	assert.Equal(t, 2, v2.Int())

	time.Sleep(150 * time.Millisecond)
	v1, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v1)
}

func TestAdapterMemory_SetIfNotExist(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	ok, err := cache.SetIfNotExist(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = cache.SetIfNotExist(ctx, "key1", "value2", 0)
	assert.NoError(t, err)
	assert.False(t, ok)

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v.String())
}

func TestAdapterMemory_SetIfNotExist_WithExpired(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	ok, err := cache.SetIfNotExist(ctx, "key1", "value1", 50*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, ok)

	time.Sleep(100 * time.Millisecond)

	ok, err = cache.SetIfNotExist(ctx, "key1", "value2", 0)
	assert.NoError(t, err)
	assert.True(t, ok)

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "value2", v.String())
}

func TestAdapterMemory_SetIfNotExistFunc(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	called := false
	f := func(ctx context.Context) (interface{}, error) {
		called = true
		return "value1", nil
	}

	ok, err := cache.SetIfNotExistFunc(ctx, "key1", f, 0)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, called)

	called = false
	ok, err = cache.SetIfNotExistFunc(ctx, "key1", f, 0)
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, called)

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v.String())
}

func TestAdapterMemory_SetIfNotExistFuncLock(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	called := false
	f := func(ctx context.Context) (interface{}, error) {
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

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", v.String())
}

func TestAdapterMemory_GetOrSet(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	v, err := cache.GetOrSet(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	assert.Equal(t, "value1", v.String())

	v, err = cache.GetOrSet(ctx, "key1", "value2", 0)
	assert.NoError(t, err)
	assert.Equal(t, "value1", v.String())
}

func TestAdapterMemory_GetOrSetFunc(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	called := false
	f := func(ctx context.Context) (interface{}, error) {
		called = true
		return "value1", nil
	}

	v, err := cache.GetOrSetFunc(ctx, "key1", f, 0)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "value1", v.String())

	called = false
	v, err = cache.GetOrSetFunc(ctx, "key1", f, 0)
	assert.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, "value1", v.String())
}

func TestAdapterMemory_GetOrSetFuncLock(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	called := false
	f := func(ctx context.Context) (interface{}, error) {
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
}

func TestAdapterMemory_Contains(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	ok, err := cache.Contains(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = cache.Contains(ctx, "key2")
	assert.NoError(t, err)
	assert.False(t, ok)

	time.Sleep(150 * time.Millisecond)
	ok, err = cache.Contains(ctx, "key1")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestAdapterMemory_Size(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	err = cache.Set(ctx, "key2", "value2", 0)
	assert.NoError(t, err)

	size, err := cache.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, size)
}

func TestAdapterMemory_Data(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	err = cache.Set(ctx, "key2", 2, 0)
	assert.NoError(t, err)

	data, err := cache.Data(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"key1": "value1", "key2": 2}, data)
}

func TestAdapterMemory_KeysAndValues(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	err = cache.Set(ctx, "key2", 2, 0)
	assert.NoError(t, err)

	keys, err := cache.Keys(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"key1", "key2"}, keys)

	values, err := cache.Values(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []interface{}{"value1", 2}, values)
}

func TestAdapterMemory_Update(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	oldVal, exist, err := cache.Update(ctx, "key1", "newValue")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "value1", oldVal.String())

	newVal, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "newValue", newVal.String())

	// check expiration is not changed
	time.Sleep(150 * time.Millisecond)
	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// Update non-existent key
	oldVal, exist, err = cache.Update(ctx, "key2", "value2")
	assert.NoError(t, err)
	assert.False(t, exist)
	assert.Nil(t, oldVal)
}

func TestAdapterMemory_UpdateExpire(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	oldDur, err := cache.UpdateExpire(ctx, "key1", 200*time.Millisecond)
	assert.NoError(t, err)
	assert.True(t, oldDur > 0 && oldDur <= 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)
	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)

	time.Sleep(100 * time.Millisecond)
	v, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestAdapterMemory_GetExpire(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 100*time.Millisecond)
	assert.NoError(t, err)

	d, err := cache.GetExpire(ctx, "key1")
	assert.NoError(t, err)
	assert.True(t, d > 0 && d <= 100*time.Millisecond)

	err = cache.Set(ctx, "key2", "value2", 0)
	assert.NoError(t, err)

	d, err = cache.GetExpire(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), d)
}

func TestAdapterMemory_Remove(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)

	lastVal, err := cache.Remove(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, lastVal)
	assert.Equal(t, "value1", lastVal.String())

	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	v, err = cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.NotNil(t, v)
}

func TestAdapterMemory_Clear(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)

	err := cache.Clear(ctx)
	assert.NoError(t, err)

	size, err := cache.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
}

func TestAdapterMemory_LRU(t *testing.T) {
	cache := mcache.NewAdapterMemory(2) // capacity is 2
	ctx := context.Background()

	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)

	// Access key1 to make it recently used
	_, _ = cache.Get(ctx, "key1")

	_ = cache.Set(ctx, "key3", "value3", 0)

	// key2 should be evicted
	v1, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v1)

	v2, err := cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Nil(t, v2)

	v3, err := cache.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.NotNil(t, v3)

	// Add another key, key1 should be evicted
	_ = cache.Set(ctx, "key4", "value4", 0)
	v1, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v1)
}

func TestAdapterMemory_Close(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	err := cache.Close(ctx)
	assert.NoError(t, err)
	// can be closed multiple times
	err = cache.Close(ctx)
	assert.NoError(t, err)
}

func TestAdapterMemory_Get_ReturnNil(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	v, err := cache.Get(ctx, "non-existent-key")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestAdapterMemory_Remove_NonExistent(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	lastValue, err := cache.Remove(ctx, "non-existent-key")
	assert.NoError(t, err)
	assert.Nil(t, lastValue)
}

func TestAdapterMemory_Update_NonExistent(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	oldValue, exist, err := cache.Update(ctx, "non-existent-key", "value")
	assert.NoError(t, err)
	assert.False(t, exist)
	assert.Nil(t, oldValue)
}

func TestAdapterMemory_UpdateExpire_NonExistent(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	oldDuration, err := cache.UpdateExpire(ctx, "non-existent-key", time.Second)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(-1), oldDuration)
}

func TestAdapterMemory_GetExpire_NonExistent(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	duration, err := cache.GetExpire(ctx, "non-existent-key")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(-1), duration)
}

func TestAdapterMemory_Set_NilValue(t *testing.T) {
	// In current implementation, setting nil does not delete the key
	// It just sets the value to nil
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()

	err := cache.Set(ctx, "key1", "value1", 0)
	assert.NoError(t, err)
	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)

	err = cache.Set(ctx, "key1", nil, 0)
	assert.NoError(t, err)
	v, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Nil(t, v.Val())
}

func TestAdapterMemory_SetIfNotExistFunc_Error(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	f := func(ctx context.Context) (interface{}, error) {
		return nil, assert.AnError
	}
	ok, err := cache.SetIfNotExistFunc(ctx, "key1", f, 0)
	assert.Error(t, err)
	assert.False(t, ok)
	v, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestAdapterMemory_GetOrSetFunc_Error(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	f := func(ctx context.Context) (interface{}, error) {
		return nil, assert.AnError
	}
	v, err := cache.GetOrSetFunc(ctx, "key1", f, 0)
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestAdapterMemory_Remove_MultipleKeys(t *testing.T) {
	cache := mcache.NewAdapterMemory()
	ctx := context.Background()
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", 2, 0)
	_ = cache.Set(ctx, "key3", "value3", 0)

	lastValue, err := cache.Remove(ctx, "key1", "key2")
	assert.NoError(t, err)
	assert.NotNil(t, lastValue)
	// The last removed value is returned, which could be either "value1" or "value2"
	// depending on map iteration order. Let's check if it's one of them.
	assert.Contains(t, []interface{}{"value1", 2}, lastValue.Val())

	v1, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v1)
	v2, err := cache.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Nil(t, v2)
	v3, err := cache.Get(ctx, "key3")
	assert.NoError(t, err)
	assert.NotNil(t, v3)
}
