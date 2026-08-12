package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/graingo/maltose/database/mredis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapterRedisDataLifecycle(t *testing.T) {
	adapter, server := newMiniRedisAdapter(t)
	ctx := context.Background()
	assert.NotNil(t, NewAdapterRedis(adapter.redis))

	require.NoError(t, adapter.Set(ctx, "plain", "value", 0))
	value, err := adapter.Get(ctx, "plain")
	require.NoError(t, err)
	assert.Equal(t, "value", value.String())
	contains, err := adapter.Contains(ctx, "plain")
	require.NoError(t, err)
	assert.True(t, contains)

	require.NoError(t, adapter.Set(ctx, "expiring", "value", time.Minute))
	expire, err := adapter.GetExpire(ctx, "expiring")
	require.NoError(t, err)
	assert.Equal(t, time.Minute, expire)
	server.FastForward(2 * time.Minute)
	value, err = adapter.Get(ctx, "expiring")
	require.NoError(t, err)
	assert.Nil(t, value)

	require.NoError(t, adapter.Set(ctx, "deleted", "value", 0))
	require.NoError(t, adapter.Set(ctx, "deleted", nil, 0))
	value, err = adapter.Get(ctx, "deleted")
	require.NoError(t, err)
	assert.Nil(t, value)

	require.NoError(t, adapter.SetMap(ctx, nil, 0))
	require.NoError(t, adapter.SetMap(ctx, map[string]interface{}{"one": 1, "two": "2"}, 0))
	require.NoError(t, adapter.SetMap(ctx, map[string]interface{}{"temporary": "value"}, time.Minute))
	data, err := adapter.Data(ctx)
	require.NoError(t, err)
	assert.Equal(t, "1", data["one"])
	assert.Equal(t, "2", data["two"])

	keys, err := adapter.Keys(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"plain", "one", "two", "temporary"}, keys)
	values, err := adapter.Values(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []interface{}{"value", "1", "2", "value"}, values)
	size, err := adapter.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 4, size)

	last, err := adapter.Remove(ctx, "one", "two")
	require.NoError(t, err)
	assert.Equal(t, "2", last.String())
	last, err = adapter.Remove(ctx)
	require.NoError(t, err)
	assert.Nil(t, last)

	require.NoError(t, adapter.SetMap(ctx, map[string]interface{}{"plain": nil, "temporary": nil}, -1))
	require.NoError(t, adapter.Clear(ctx))
	size, err = adapter.Size(ctx)
	require.NoError(t, err)
	assert.Zero(t, size)
	assert.NoError(t, adapter.Close(ctx))
}

func TestAdapterRedisConditionalAndComputedValues(t *testing.T) {
	adapter, _ := newMiniRedisAdapter(t)
	ctx := context.Background()

	set, err := adapter.SetIfNotExist(ctx, "key", "first", 0)
	require.NoError(t, err)
	assert.True(t, set)
	set, err = adapter.SetIfNotExist(ctx, "key", "second", 0)
	require.NoError(t, err)
	assert.False(t, set)
	set, err = adapter.SetIfNotExist(ctx, "key", nil, -1)
	require.NoError(t, err)
	assert.True(t, set)

	called := 0
	compute := func(context.Context) (interface{}, error) {
		called++
		return "computed", nil
	}
	set, err = adapter.SetIfNotExistFunc(ctx, "computed", compute, 0)
	require.NoError(t, err)
	assert.True(t, set)
	set, err = adapter.SetIfNotExistFunc(ctx, "computed", compute, 0)
	require.NoError(t, err)
	assert.False(t, set)
	assert.Equal(t, 1, called)

	expectedErr := errors.New("compute failed")
	set, err = adapter.SetIfNotExistFunc(ctx, "failed", func(context.Context) (interface{}, error) {
		return nil, expectedErr
	}, 0)
	assert.False(t, set)
	assert.ErrorIs(t, err, expectedErr)

	value, err := adapter.GetOrSet(ctx, "direct", "first", 0)
	require.NoError(t, err)
	assert.Equal(t, "first", value.String())
	value, err = adapter.GetOrSet(ctx, "direct", "second", 0)
	require.NoError(t, err)
	assert.Equal(t, "first", value.String())

	value, err = adapter.GetOrSetFunc(ctx, "function", compute, 0)
	require.NoError(t, err)
	assert.Equal(t, "computed", value.String())
	previousCalls := called
	value, err = adapter.GetOrSetFunc(ctx, "function", compute, 0)
	require.NoError(t, err)
	assert.Equal(t, "computed", value.String())
	assert.Equal(t, previousCalls, called)
	_, err = adapter.GetOrSetFunc(ctx, "function-error", func(context.Context) (interface{}, error) {
		return nil, expectedErr
	}, 0)
	assert.ErrorIs(t, err, expectedErr)
}

func TestAdapterRedisLocksAndContextCancellation(t *testing.T) {
	adapter, _ := newMiniRedisAdapter(t,
		WithLockRetryInterval(time.Millisecond),
		WithLockWaitTimeout(50*time.Millisecond),
	)
	ctx := context.Background()

	called := 0
	compute := func(context.Context) (interface{}, error) {
		called++
		return "locked", nil
	}
	set, err := adapter.SetIfNotExistFuncLock(ctx, "conditional", compute, 0)
	require.NoError(t, err)
	assert.True(t, set)
	set, err = adapter.SetIfNotExistFuncLock(ctx, "conditional", compute, 0)
	require.NoError(t, err)
	assert.False(t, set)

	value, err := adapter.GetOrSetFuncLock(ctx, "single-flight", compute, 0)
	require.NoError(t, err)
	assert.Equal(t, "locked", value.String())
	value, err = adapter.GetOrSetFuncLock(ctx, "single-flight", compute, 0)
	require.NoError(t, err)
	assert.Equal(t, "locked", value.String())

	lockKey := cacheLockKey(adapter.cacheKey("blocked"))
	require.NoError(t, adapter.redis.SetEX(ctx, lockKey, "another-owner", time.Minute))
	canceled, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, err = adapter.GetOrSetFuncLock(canceled, "blocked", compute, 0)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	require.NoError(t, adapter.redis.SetEX(ctx, lockKey, "owner-a", time.Minute))
	require.NoError(t, adapter.releaseLock(ctx, lockKey, "owner-b"))
	exists, err := adapter.redis.Exists(ctx, lockKey)
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)
	require.NoError(t, adapter.releaseLock(ctx, lockKey, "owner-a"))
}

func TestAdapterRedisExpirationAndUpdate(t *testing.T) {
	adapter, _ := newMiniRedisAdapter(t)
	ctx := context.Background()

	old, exists, err := adapter.Update(ctx, "missing", "value")
	require.NoError(t, err)
	assert.False(t, exists)
	assert.Nil(t, old)

	require.NoError(t, adapter.Set(ctx, "key", "old", time.Minute))
	old, exists, err = adapter.Update(ctx, "key", "new")
	require.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, "old", old.String())

	oldDuration, err := adapter.UpdateExpire(ctx, "missing", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), oldDuration)
	require.NoError(t, adapter.Set(ctx, "persistent", "value", 0))
	oldDuration, err = adapter.UpdateExpire(ctx, "persistent", time.Minute)
	require.NoError(t, err)
	assert.Zero(t, oldDuration)
	oldDuration, err = adapter.UpdateExpire(ctx, "persistent", 0)
	require.NoError(t, err)
	assert.Positive(t, oldDuration)
	_, err = adapter.UpdateExpire(ctx, "persistent", -1)
	require.NoError(t, err)

	expire, err := adapter.GetExpire(ctx, "missing")
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), expire)
	require.NoError(t, adapter.Set(ctx, "persistent", "value", 0))
	expire, err = adapter.GetExpire(ctx, "persistent")
	require.NoError(t, err)
	assert.Zero(t, expire)
}

func TestAdapterRedisPrefixScopesCollectionOperations(t *testing.T) {
	adapter, _ := newMiniRedisAdapter(t, WithKeyPrefix("tenant[1]:"))
	ctx := context.Background()

	require.NoError(t, adapter.redis.Set(ctx, "outside", "preserved"))
	require.NoError(t, adapter.redis.Set(ctx, "tenant1:glob-match", "preserved"))
	require.NoError(t, adapter.Set(ctx, "one", "1", 0))
	require.NoError(t, adapter.Set(ctx, "two", "2", 0))

	keys, err := adapter.Keys(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"one", "two"}, keys)
	size, err := adapter.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, size)

	require.NoError(t, adapter.Clear(ctx))
	exists, err := adapter.redis.Exists(ctx, "outside", "tenant1:glob-match")
	require.NoError(t, err)
	assert.Equal(t, int64(2), exists)
}

func newMiniRedisAdapter(t *testing.T, options ...Option) (*AdapterRedis, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client, err := mredis.New(&mredis.Config{Address: server.Addr()})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	adapter := NewAdapterRedisWithOptions(client, options...).(*AdapterRedis)
	return adapter, server
}
