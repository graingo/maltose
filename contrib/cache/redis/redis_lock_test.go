package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockOptions(t *testing.T) {
	adapter := NewAdapterRedis(
		nil,
		WithKeyPrefix("app:cache:"),
		WithLockTTL(3*time.Second),
		WithLockRetryInterval(20*time.Millisecond),
		WithLockWaitTimeout(5*time.Second),
	).(*AdapterRedis)

	assert.Equal(t, "app:cache:", adapter.keyPrefix)
	assert.Equal(t, 3*time.Second, adapter.lockTTL)
	assert.Equal(t, 20*time.Millisecond, adapter.lockRetryInterval)
	assert.Equal(t, 5*time.Second, adapter.lockWaitTimeout)
}

func TestCacheKeyMapping(t *testing.T) {
	adapter := NewAdapterRedis(nil, WithKeyPrefix("app:cache:")).(*AdapterRedis)

	assert.Equal(t, "app:cache:user:1", adapter.cacheKey("user:1"))
	assert.Equal(t, "user:1", adapter.logicalKey("app:cache:user:1"))
}

func TestNewLockTokenReturnsUniqueTokens(t *testing.T) {
	first, err := newLockToken()
	require.NoError(t, err)
	second, err := newLockToken()
	require.NoError(t, err)

	assert.Len(t, first, 32)
	assert.Len(t, second, 32)
	assert.NotEqual(t, first, second)
}
