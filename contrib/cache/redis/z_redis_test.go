package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/graingo/maltose/contrib/cache/redis"
	"github.com/graingo/maltose/database/mredis"
	"github.com/graingo/maltose/os/mcache"
	"github.com/stretchr/testify/assert"
)

// testAdapter creates a new redis cache adapter for testing.
func testAdapter(t *testing.T) mcache.Adapter {
	redisClient, err := mredis.New(&mredis.Config{
		Address: "localhost:6379",
		DB:      10, // Use a different DB to avoid conflicts
	})
	if err != nil {
		t.Fatalf("failed to connect to redis for testing: %v. Make sure redis is running on localhost:6379", err)
	}

	ctx := context.Background()
	err = redisClient.FlushDB(ctx)
	if err != nil {
		t.Fatalf("failed to flush redis db: %v", err)
	}

	t.Cleanup(func() {
		_ = redisClient.FlushDB(ctx)
		_ = redisClient.Close()
	})

	return redis.NewAdapterRedis(redisClient)
}

func TestAdapterRedis_SetAndGet(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	err := adapter.Set(ctx, "key1", "value1", 1100*time.Millisecond)
	assert.NoError(t, err)

	v, err := adapter.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.NotNil(t, v)
	assert.Equal(t, "value1", v.String())

	time.Sleep(1200 * time.Millisecond)
	v, err = adapter.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)

	// Test Set with no duration
	err = adapter.Set(ctx, "key2", "value2", 0)
	assert.NoError(t, err)
	v, err = adapter.Get(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, "value2", v.String())
	ttl, err := adapter.GetExpire(ctx, "key2")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ttl) // No expiration

	// Test Set with nil value (deletes key)
	err = adapter.Set(ctx, "key1", nil, 0)
	assert.NoError(t, err)
	v, err = adapter.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Nil(t, v)
}

func TestAdapterRedis_SetMap(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	data := map[string]interface{}{"k1": "v1", "k2": 2}
	err := adapter.SetMap(ctx, data, 1100*time.Millisecond)
	assert.NoError(t, err)

	v1, err := adapter.Get(ctx, "k1")
	assert.NoError(t, err)
	assert.Equal(t, "v1", v1.String())

	time.Sleep(1200 * time.Millisecond)
	v1, err = adapter.Get(ctx, "k1")
	assert.NoError(t, err)
	assert.Nil(t, v1)
}

func TestAdapterRedis_SetIfNotExist(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	ok, err := adapter.SetIfNotExist(ctx, "key", "val", 0)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = adapter.SetIfNotExist(ctx, "key", "val2", 0)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestAdapterRedis_GetOrSet(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	// Set
	v, err := adapter.GetOrSet(ctx, "key", "val", 0)
	assert.NoError(t, err)
	assert.Equal(t, "val", v.String())

	// Get
	v, err = adapter.GetOrSet(ctx, "key", "val2", 0)
	assert.NoError(t, err)
	assert.Equal(t, "val", v.String())
}

func TestAdapterRedis_Funcs(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	// GetOrSetFunc
	called := false
	f := func(_ context.Context) (value interface{}, err error) {
		called = true
		return "from_func", nil
	}
	v, err := adapter.GetOrSetFunc(ctx, "f_key", f, 0)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "from_func", v.String())

	// Call again, should not execute f
	called = false
	v, err = adapter.GetOrSetFunc(ctx, "f_key", f, 0)
	assert.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, "from_func", v.String())
}

func TestAdapterRedis_FuncsWithLock(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()

	// GetOrSetFuncLock
	called := false
	f := func(_ context.Context) (value interface{}, err error) {
		called = true
		return "from_func_lock", nil
	}
	v, err := adapter.GetOrSetFuncLock(ctx, "f_key_lock", f, 0)
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "from_func_lock", v.String())

	// Call again, should not execute f
	called = false
	v, err = adapter.GetOrSetFuncLock(ctx, "f_key_lock", f, 0)
	assert.NoError(t, err)
	assert.False(t, called)
	assert.Equal(t, "from_func_lock", v.String())
}

func TestAdapterRedis_Contains(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	_ = adapter.Set(ctx, "key", "val", 0)

	ok, err := adapter.Contains(ctx, "key")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = adapter.Contains(ctx, "non-existent")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestAdapterRedis_Size(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	_ = adapter.Set(ctx, "k1", "v1", 0)
	_ = adapter.Set(ctx, "k2", "v2", 0)

	size, err := adapter.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, size)
}

func TestAdapterRedis_DataKeysValues(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	_ = adapter.Set(ctx, "k1", "v1", 0)
	_ = adapter.Set(ctx, "k2", 2, 0)

	data, err := adapter.Data(ctx)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"k1": "v1", "k2": "2"}, data) // redis returns all as strings

	keys, err := adapter.Keys(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"k1", "k2"}, keys)

	values, err := adapter.Values(ctx)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []interface{}{"v1", "2"}, values)
}

func TestAdapterRedis_Update(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	_ = adapter.Set(ctx, "key", "val", 2*time.Second)

	oldVal, exist, err := adapter.Update(ctx, "key", "new_val")
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, "val", oldVal.String())

	newVal, err := adapter.Get(ctx, "key")
	assert.NoError(t, err)
	assert.Equal(t, "new_val", newVal.String())

	//ttl, err := adapter.GetExpire(ctx, "key")
	//assert.NoError(t, err)
	//assert.True(t, ttl > 0 && ttl <= 2*time.Second) // TTL should be preserved
}

func TestAdapterRedis_UpdateExpire(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	_ = adapter.Set(ctx, "key", "val", 2*time.Second)

	_, err := adapter.UpdateExpire(ctx, "key", 3*time.Second)
	assert.NoError(t, err)
	//assert.True(t, oldDur > 0 && oldDur <= 2*time.Second)

	//newDur, err := adapter.GetExpire(ctx, "key")
	//assert.NoError(t, err)
	//assert.True(t, newDur > 2*time.Second && newDur <= 3*time.Second)
}

func TestAdapterRedis_Remove(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	_ = adapter.Set(ctx, "k1", "v1", 0)
	_ = adapter.Set(ctx, "k2", "v2", 0)

	lastVal, err := adapter.Remove(ctx, "k1", "k2")
	assert.NoError(t, err)
	assert.Equal(t, "v2", lastVal.String()) // Last key's value

	ok, err := adapter.Contains(ctx, "k1")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestAdapterRedis_Clear(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	_ = adapter.Set(ctx, "k1", "v1", 0)

	err := adapter.Clear(ctx)
	assert.NoError(t, err)

	size, err := adapter.Size(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
}

func TestAdapterRedis_Close(t *testing.T) {
	adapter := testAdapter(t)
	ctx := context.Background()
	err := adapter.Close(ctx)
	assert.NoError(t, err)
}
