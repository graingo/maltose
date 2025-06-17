package mredis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/graingo/maltose/database/mredis"
	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
)

// testClient is a helper to get a redis client for testing.
// It assumes a redis server is running on localhost:6379, and uses DB 9.
func testClient(t *testing.T) *mredis.Redis {
	client, err := mredis.New(&mredis.Config{
		Address: "localhost:6379",
		DB:      9,
	})
	if err != nil {
		t.Fatalf("failed to connect to redis for testing: %v", err)
	}

	ctx := context.Background()
	err = client.FlushDB(ctx)
	if err != nil {
		t.Fatalf("failed to flush redis db: %v", err)
	}

	t.Cleanup(func() {
		_ = client.FlushDB(ctx)
		_ = client.Close()
	})

	return client
}

func TestNew(t *testing.T) {
	// Test with valid config
	client, err := mredis.New(&mredis.Config{Address: "localhost:6379"})
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NoError(t, client.Ping(context.Background()))
	assert.NoError(t, client.Close())

	// Test with nil config
	client, err = mredis.New(nil)
	assert.Error(t, err)
	assert.Nil(t, client)

	// Test without config
	client, err = mredis.New()
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestPing(t *testing.T) {
	client := testClient(t)
	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestGeneric(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

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
}

func TestString(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

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
}

func TestHash(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
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
}

func TestList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	key := "mylist"

	_, err := client.LPush(ctx, key, "one", "two")
	assert.NoError(t, err)

	res, err := client.RPop(ctx, key)
	assert.NoError(t, err)
	assert.Equal(t, "one", res.String())
}

func TestSet(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	key := "myset"

	_, err := client.SAdd(ctx, key, "one", "two")
	assert.NoError(t, err)

	isMember, err := client.SIsMember(ctx, key, "one")
	assert.NoError(t, err)
	assert.True(t, isMember)
}

func TestSortedSet(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	key := "myzset"
	member := mredis.Z{Score: 1, Member: "one"}

	_, err := client.ZAdd(ctx, key, member)
	assert.NoError(t, err)

	members, err := client.Client().ZRange(ctx, key, 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"one"}, members)
}

func TestInstance(t *testing.T) {
	ctx := context.Background()

	// Test a named instance with pre-set configuration
	namedCfg := &mredis.Config{Address: "localhost:6379", DB: 8}
	mredis.SetConfig("my-test-instance", namedCfg)
	defer mredis.RemoveConfig("my-test-instance")

	namedInstance := mredis.Instance("my-test-instance")
	assert.NotNil(t, namedInstance)
	err := namedInstance.Ping(ctx)
	assert.NoError(t, err)
	assert.NoError(t, namedInstance.Close())

	// Test default instance with pre-set configuration
	defaultCfgMap := map[string]any{"address": "localhost:6379", "db": 7}
	err = mredis.SetConfigByMap(defaultCfgMap)
	assert.NoError(t, err)
	defer mredis.RemoveConfig()

	defaultInstance := mredis.Instance()
	assert.NotNil(t, defaultInstance)
	err = defaultInstance.Ping(ctx)
	assert.NoError(t, err)
	assert.NoError(t, defaultInstance.Close())

	// Test getting an instance for which no configuration is set
	nilInstance := mredis.Instance("non-existent-instance")
	assert.Nil(t, nilInstance)
}

func TestClient(t *testing.T) {
	client := testClient(t)
	assert.NotNil(t, client.Client())
}

func TestRedis_LoggingHook(t *testing.T) {
	// Create a client with logger configured
	client, err := mredis.New(&mredis.Config{
		Address:       "localhost:6379",
		DB:            9,
		Logger:        mlog.New(), // Enable the logger
		SlowThreshold: 1 * time.Millisecond,
	})
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	client.FlushDB(ctx)

	// --- Test a normal command ---
	fmt.Println("--- Testing a normal command ---")
	err = client.Set(ctx, "fast_key", "value")
	assert.NoError(t, err)

	// --- Test a slow command (simulated) ---
	fmt.Println("\n--- Testing a slow command ---")
	// Using a command that might take a bit longer, or just for demonstration
	_, err = client.Keys(ctx, "*")
	assert.NoError(t, err)

	// --- Test a command that returns an error (by using wrong type) ---
	fmt.Println("\n--- Testing a command with an error ---")
	err = client.HSet(ctx, "a_string_key", map[string]interface{}{"f": "v"})
	assert.NoError(t, err) // This will set the key
	_, err = client.Client().Incr(ctx, "a_string_key").Result()
	assert.Error(t, err) // This should fail
}
