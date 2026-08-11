//go:build integration

package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/graingo/maltose/database/mredis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseLockChecksOwnership(t *testing.T) {
	client := integrationRedisClient(t)
	adapter := NewAdapterRedis(client).(*AdapterRedis)
	ctx := context.Background()
	lockKey := cacheLockKey("ownership")

	require.NoError(t, client.SetEX(ctx, lockKey, "owner-a", time.Second))
	require.NoError(t, adapter.releaseLock(ctx, lockKey, "owner-b"))
	exists, err := client.Exists(ctx, lockKey)
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	require.NoError(t, adapter.releaseLock(ctx, lockKey, "owner-a"))
	exists, err = client.Exists(ctx, lockKey)
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestGetOrSetFuncLockHonorsContext(t *testing.T) {
	client := integrationRedisClient(t)
	adapter := NewAdapterRedisWithOptions(
		client,
		WithLockRetryInterval(5*time.Millisecond),
		WithLockWaitTimeout(time.Second),
	).(*AdapterRedis)
	lockKey := cacheLockKey("blocked")
	require.NoError(t, client.SetEX(context.Background(), lockKey, "another-owner", time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	_, err := adapter.GetOrSetFuncLock(ctx, "blocked", func(context.Context) (interface{}, error) {
		return "value", nil
	}, time.Minute)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestKeyPrefixScopesGlobalOperations(t *testing.T) {
	client := integrationRedisClient(t)
	adapter := NewAdapterRedisWithOptions(client, WithKeyPrefix("app:cache:")).(*AdapterRedis)
	ctx := context.Background()

	require.NoError(t, client.Set(ctx, "outside", "preserved"))
	require.NoError(t, adapter.Set(ctx, "first", "one", 0))
	require.NoError(t, adapter.SetMap(ctx, map[string]interface{}{"second": "two", "third": "three"}, 0))

	keys, err := adapter.Keys(ctx)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"first", "second", "third"}, keys)
	size, err := adapter.Size(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, size)
	data, err := adapter.Data(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"first": "one", "second": "two", "third": "three"}, data)

	require.NoError(t, adapter.Clear(ctx))
	exists, err := client.Exists(ctx, "outside")
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)
	size, err = adapter.Size(ctx)
	require.NoError(t, err)
	assert.Zero(t, size)
}

func integrationRedisClient(t *testing.T) *mredis.Redis {
	t.Helper()
	client, err := mredis.New(&mredis.Config{Address: "localhost:6379", DB: 10})
	require.NoError(t, err)
	require.NoError(t, client.FlushDB(context.Background()))
	t.Cleanup(func() {
		_ = client.FlushDB(context.Background())
		_ = client.Close()
	})
	return client
}
