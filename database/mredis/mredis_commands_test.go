package mredis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisCommandGroups(t *testing.T) {
	client, server := newMiniRedisClient(t)
	ctx := context.Background()
	require.NoError(t, client.Ping(ctx))

	t.Run("generic and string commands", func(t *testing.T) {
		require.NoError(t, client.Set(ctx, "plain", "value"))
		value, err := client.Get(ctx, "plain")
		require.NoError(t, err)
		assert.Equal(t, "value", value.String())

		exists, err := client.Exists(ctx, "plain", "missing")
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
		keys, err := client.Keys(ctx, "pl*")
		require.NoError(t, err)
		assert.Equal(t, []string{"plain"}, keys)

		require.NoError(t, client.SetEX(ctx, "expiring", "value", time.Minute))
		ttl, err := client.TTL(ctx, "expiring")
		require.NoError(t, err)
		assert.Positive(t, ttl)
		persisted, err := client.Persist(ctx, "expiring")
		require.NoError(t, err)
		assert.True(t, persisted)
		expired, err := client.Expire(ctx, "expiring", time.Minute)
		require.NoError(t, err)
		assert.True(t, expired)

		set, err := client.SetNX(ctx, "unique", "first", 0)
		require.NoError(t, err)
		assert.True(t, set)
		set, err = client.SetNX(ctx, "unique", "second", 0)
		require.NoError(t, err)
		assert.False(t, set)

		require.NoError(t, client.MSet(ctx, map[string]interface{}{"one": 1, "two": "2"}))
		values, err := client.MGet(ctx, "one", "two", "missing")
		require.NoError(t, err)
		require.Len(t, values, 3)
		assert.Equal(t, "1", values[0].String())
		assert.Equal(t, "2", values[1].String())
		assert.True(t, values[2].IsNil())

		size, err := client.DBSize(ctx)
		require.NoError(t, err)
		assert.Positive(t, size)
		deleted, err := client.Del(ctx, "plain", "missing")
		require.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
	})

	t.Run("hash list set and sorted set commands", func(t *testing.T) {
		require.NoError(t, client.HSet(ctx, "hash", map[string]interface{}{"field": "value"}))
		value, err := client.HGet(ctx, "hash", "field")
		require.NoError(t, err)
		assert.Equal(t, "value", value.String())
		value, err = client.HGet(ctx, "hash", "missing")
		require.NoError(t, err)
		assert.Nil(t, value)

		length, err := client.LPush(ctx, "list", "one", "two")
		require.NoError(t, err)
		assert.Equal(t, int64(2), length)
		value, err = client.RPop(ctx, "list")
		require.NoError(t, err)
		assert.Equal(t, "one", value.String())
		_, err = client.RPop(ctx, "missing-list")
		require.NoError(t, err)

		added, err := client.SAdd(ctx, "set", "one", "two")
		require.NoError(t, err)
		assert.Equal(t, int64(2), added)
		member, err := client.SIsMember(ctx, "set", "one")
		require.NoError(t, err)
		assert.True(t, member)

		added, err = client.ZAdd(ctx, "sorted", Z{Score: 1, Member: "one"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), added)
		score, err := client.ZScore(ctx, "sorted", "one")
		require.NoError(t, err)
		assert.Equal(t, float64(1), score)
	})

	t.Run("missing values and type errors", func(t *testing.T) {
		value, err := client.Get(ctx, "missing")
		require.NoError(t, err)
		assert.Nil(t, value)

		require.NoError(t, client.Set(ctx, "wrong-type", "value"))
		_, err = client.HGet(ctx, "wrong-type", "field")
		assert.Error(t, err)
		_, err = client.RPop(ctx, "wrong-type")
		assert.Error(t, err)
		_, err = client.ZScore(ctx, "sorted", "missing")
		assert.ErrorIs(t, err, redis.Nil)
	})

	server.FastForward(2 * time.Minute)
	require.NoError(t, client.FlushDB(ctx))
	size, err := client.DBSize(ctx)
	require.NoError(t, err)
	assert.Zero(t, size)
}

func newMiniRedisClient(t *testing.T) (*Redis, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client, err := New(&Config{Address: server.Addr()})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	return client, server
}
