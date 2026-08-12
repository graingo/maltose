package mredis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMergeConfigPreservesDefaults(t *testing.T) {
	config := mergeConfig(&Config{DB: 4})

	assert.Equal(t, "127.0.0.1:6379", config.Address)
	assert.Equal(t, 4, config.DB)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.Equal(t, 3*time.Second, config.ReadTimeout)
	assert.Equal(t, 3*time.Second, config.WriteTimeout)
}

func TestMergeConfigAppliesOverrides(t *testing.T) {
	config := mergeConfig(&Config{
		Address:      "redis.internal:6380",
		PoolSize:     32,
		DialTimeout:  time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 4 * time.Second,
	})

	assert.Equal(t, "redis.internal:6380", config.Address)
	assert.Equal(t, 32, config.PoolSize)
	assert.Equal(t, time.Second, config.DialTimeout)
	assert.Equal(t, 2*time.Second, config.ReadTimeout)
	assert.Equal(t, 4*time.Second, config.WriteTimeout)
}
