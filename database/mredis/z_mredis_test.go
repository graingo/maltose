package mredis_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/graingo/maltose/database/mredis"
	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testClient is a helper to get a redis client for testing.
// It assumes a redis server is running on localhost:6379, and uses DB 9.
func testClient(t *testing.T) *mredis.Redis {
	client, err := mredis.New(&mredis.Config{
		Address: "localhost:6379",
		DB:      9,
	})
	// Use require here because a failed connection is a fatal error for the test.
	require.NoError(t, err, "failed to connect to redis for testing")

	ctx := context.Background()
	err = client.FlushDB(ctx)
	require.NoError(t, err, "failed to flush redis db")

	t.Cleanup(func() {
		_ = client.FlushDB(ctx)
		_ = client.Close()
	})

	return client
}

func TestRedis(t *testing.T) {
	t.Run("connection_and_config", func(t *testing.T) {
		t.Run("new_client", func(t *testing.T) {
			// Test with valid config
			client, err := mredis.New(&mredis.Config{Address: "localhost:6379"})
			assert.NoError(t, err)
			require.NotNil(t, client)
			assert.NoError(t, client.Ping(context.Background()))
			assert.NoError(t, client.Close())

			// Test with nil config, should fall back to default config.
			client, err = mredis.New(nil)
			assert.NoError(t, err)
			require.NotNil(t, client)
			assert.NoError(t, client.Ping(context.Background()))
			assert.NoError(t, client.Close())

			// Test with no config, should fall back to default config.
			client, err = mredis.New()
			assert.NoError(t, err)
			require.NotNil(t, client)
			assert.NoError(t, client.Ping(context.Background()))
			assert.NoError(t, client.Close())
		})

		t.Run("instance_management", func(t *testing.T) {
			ctx := context.Background()

			// Test a named instance with pre-set configuration
			namedCfg := &mredis.Config{Address: "localhost:6379", DB: 8}
			mredis.SetConfig("my-test-instance", namedCfg)
			defer mredis.RemoveConfig("my-test-instance")

			namedInstance := mredis.Instance("my-test-instance")
			require.NotNil(t, namedInstance)
			err := namedInstance.Ping(ctx)
			assert.NoError(t, err)
			assert.NoError(t, namedInstance.Close())

			// Test default instance with pre-set configuration via map
			defaultCfgMap := map[string]any{"address": "localhost:6379", "db": 7}
			err = mredis.SetConfigByMap(defaultCfgMap)
			require.NoError(t, err)
			defer mredis.RemoveConfig()

			defaultInstance := mredis.Instance()
			require.NotNil(t, defaultInstance)
			err = defaultInstance.Ping(ctx)
			assert.NoError(t, err)
			assert.NoError(t, defaultInstance.Close())

			// Test getting an instance for which no configuration is set
			nilInstance := mredis.Instance("non-existent-instance")
			assert.Nil(t, nilInstance)
		})
	})

	t.Run("command_operations", func(t *testing.T) {
		client := testClient(t)
		ctx := context.Background()

		t.Run("generic_commands", func(t *testing.T) {
			res, err := client.Exists(ctx, "key")
			assert.NoError(t, err)
			assert.Equal(t, int64(0), res)

			err = client.Set(ctx, "key", "val")
			assert.NoError(t, err)

			res, err = client.Exists(ctx, "key")
			assert.NoError(t, err)
			assert.Equal(t, int64(1), res)

			ttl, err := client.TTL(ctx, "key")
			assert.NoError(t, err)
			assert.Equal(t, time.Duration(-1), ttl) // No expiration

			_, err = client.Expire(ctx, "key", 10*time.Second)
			assert.NoError(t, err)

			ttl, err = client.TTL(ctx, "key")
			assert.NoError(t, err)
			assert.True(t, ttl > 0 && ttl <= 10*time.Second)

			_, err = client.Del(ctx, "key")
			assert.NoError(t, err)
		})

		t.Run("string_commands", func(t *testing.T) {
			err := client.Set(ctx, "key", "val")
			assert.NoError(t, err)

			val, err := client.Get(ctx, "key")
			assert.NoError(t, err)
			assert.Equal(t, "val", val.String())

			res, err := client.Client().GetSet(ctx, "key", "newVal").Result()
			assert.NoError(t, err)
			assert.Equal(t, "val", res)

			newVal, err := client.Get(ctx, "key")
			assert.NoError(t, err)
			assert.Equal(t, "newVal", newVal.String())
		})

		t.Run("hash_commands", func(t *testing.T) {
			key := "myhash"
			field := "f1"
			value := "v1"

			err := client.HSet(ctx, key, map[string]interface{}{field: value})
			assert.NoError(t, err)

			res, err := client.Client().HGet(ctx, key, field).Result()
			assert.NoError(t, err)
			assert.Equal(t, value, res)

			all, err := client.Client().HGetAll(ctx, key).Result()
			assert.NoError(t, err)
			assert.Equal(t, map[string]string{field: value}, all)
		})

		t.Run("list_commands", func(t *testing.T) {
			key := "mylist"
			_, err := client.LPush(ctx, key, "one", "two")
			assert.NoError(t, err)

			res, err := client.RPop(ctx, key)
			assert.NoError(t, err)
			assert.Equal(t, "one", res.String())
		})

		t.Run("set_commands", func(t *testing.T) {
			key := "myset"
			_, err := client.SAdd(ctx, key, "one", "two")
			assert.NoError(t, err)

			isMember, err := client.SIsMember(ctx, key, "one")
			assert.NoError(t, err)
			assert.True(t, isMember)
		})

		t.Run("sorted_set_commands", func(t *testing.T) {
			key := "myzset"
			member := mredis.Z{Score: 1, Member: "one"}

			_, err := client.ZAdd(ctx, key, member)
			assert.NoError(t, err)

			members, err := client.Client().ZRange(ctx, key, 0, -1).Result()
			assert.NoError(t, err)
			assert.Equal(t, []string{"one"}, members)
		})
	})

	t.Run("hooks", func(t *testing.T) {
		t.Run("logging_hook", func(t *testing.T) {
			// Configure mlog to write to a buffer for inspection.
			var logBuffer bytes.Buffer
			logger := mlog.New(&mlog.Config{
				Writer: &logBuffer,
				Stdout: false,
				Level:  mlog.DebugLevel,
				Format: "json", // JSON is easier to parse in tests.
			})

			// Create a client with our buffered logger.
			client, err := mredis.New(&mredis.Config{
				Address:       "localhost:6379",
				DB:            9,
				Logger:        logger,
				SlowThreshold: 1 * time.Millisecond,
			})
			require.NoError(t, err)
			defer client.Close()

			ctx := context.Background()
			client.FlushDB(ctx)

			// 1. Test a normal command
			err = client.Set(ctx, "fast_key", "value")
			assert.NoError(t, err)
			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, `"cmd":"set"`, "should log the normal command")
			assert.NotContains(t, logOutput, "SLOW", "normal command should not be marked as slow")
			logBuffer.Reset() // Clear buffer for next test

			// 2. Test a command that returns an error
			// Set a key so we can try an invalid operation on it.
			err = client.Set(ctx, "a_string_key", "value")
			require.NoError(t, err)
			_, err = client.Client().Incr(ctx, "a_string_key").Result()
			assert.Error(t, err)

			logOutput = logBuffer.String()
			assert.Contains(t, logOutput, `"cmd":"incr"`, "should log the error command")
			assert.Contains(t, logOutput, `"error":"ERR value is not an integer`, "should log the redis error")
			logBuffer.Reset()

			// 3. Test a slow command (simulated by setting a very low threshold)
			client.SetSlowThreshold(time.Nanosecond) // Use a very small positive duration to trigger the slow log.
			_, err = client.Keys(ctx, "*")
			assert.NoError(t, err)
			logOutput = logBuffer.String()
			assert.Contains(t, logOutput, `"cmd":"keys"`, "should log the slow command")
			assert.Contains(t, logOutput, `"level":"warn"`, "slow command should be logged at warn level")
			assert.Contains(t, logOutput, `"msg":"redis command slow"`, "slow command should have the correct message")
		})
	})
}
