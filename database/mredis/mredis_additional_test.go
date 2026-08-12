package mredis

import (
	"context"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mlog"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type passthroughHook struct{}

func (passthroughHook) DialHook(next redis.DialHook) redis.DialHook          { return next }
func (passthroughHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook { return next }
func (passthroughHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestConfigMutationCloningAndOverrides(t *testing.T) {
	hook := passthroughHook{}
	logger := mlog.New()
	config := mergeConfig(&Config{
		Address:         "redis.internal:6380",
		DB:              2,
		User:            "user",
		Password:        "secret",
		MasterName:      "master",
		MinIdleConns:    1,
		MaxIdleConns:    2,
		MaxRetries:      3,
		PoolSize:        4,
		MinRetryBackoff: time.Millisecond,
		MaxRetryBackoff: 2 * time.Millisecond,
		DialTimeout:     3 * time.Millisecond,
		ReadTimeout:     4 * time.Millisecond,
		WriteTimeout:    5 * time.Millisecond,
		PoolTimeout:     6 * time.Millisecond,
		ConnMaxIdleTime: 7 * time.Millisecond,
		SlowThreshold:   8 * time.Millisecond,
		Logger:          logger,
		Hooks:           []Hook{hook},
	})

	assert.Equal(t, "redis.internal:6380", config.Address)
	assert.Equal(t, 2, config.DB)
	assert.Equal(t, 4, config.PoolSize)
	assert.Equal(t, 8*time.Millisecond, config.SlowThreshold)
	assert.Same(t, logger, config.Logger)
	assert.Len(t, config.Hooks, 1)

	cloned := cloneConfig(config)
	cloned.Hooks = append(cloned.Hooks, hook)
	assert.Len(t, config.Hooks, 1)
	assert.Nil(t, cloned.loggerHook)
	assert.NotNil(t, cloneConfig(nil))

	require.NoError(t, config.SetConfigWithMap(map[string]any{"address": "localhost:6379", "db": 9}))
	assert.Equal(t, 9, config.DB)
	config.SetLogger(nil)
	assert.NotNil(t, config.Logger)
	config.AddHook(hook)
	assert.Len(t, config.Hooks, 2)
}

func TestClientLifecycleWithoutRedisServer(t *testing.T) {
	config := &Config{
		Address:      "127.0.0.1:1",
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
		MaxRetries:   1,
	}
	config.SetLogger(mlog.New())
	client, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, client.Client())

	client.AddHook(passthroughHook{})
	client.SetSlowThreshold(time.Nanosecond)
	assert.Error(t, client.Ping(context.Background()))
	assert.NoError(t, client.Close())
}

func TestConfigRegistryAndInstanceInvalidation(t *testing.T) {
	const name = "additional-config"
	t.Cleanup(func() { RemoveConfig(name) })

	require.NoError(t, SetConfigByMap(map[string]any{"address": "127.0.0.1:1", "db": 3}, name))
	config, ok := GetConfig(name)
	require.True(t, ok)
	assert.Equal(t, 3, config.DB)

	config.DB = 8
	stored, ok := GetConfig(name)
	require.True(t, ok)
	assert.Equal(t, 3, stored.DB, "registry values must be returned as clones")
	first := Instance(name)
	require.NotNil(t, first)

	SetConfig(name, &Config{Address: "127.0.0.1:2"})
	second := Instance(name)
	require.NotNil(t, second)
	assert.NotSame(t, first, second)

	RemoveConfig(name)
	_, ok = GetConfig(name)
	assert.False(t, ok)
	assert.Nil(t, Instance(name))

	_, err := ConfigFromMap(map[string]any{"db": []int{1}})
	assert.Error(t, err)
}

var _ Hook = passthroughHook{}
