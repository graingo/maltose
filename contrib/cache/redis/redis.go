package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/database/mredis"
	"github.com/graingo/maltose/os/mcache"
	"github.com/redis/go-redis/v9"
)

const (
	defaultLockTTL           = 10 * time.Second
	defaultLockRetryInterval = 50 * time.Millisecond
	defaultLockWaitTimeout   = 10 * time.Second
	lockReleaseTimeout       = time.Second
	lockKeySuffix            = ":maltose:lock"
)

const releaseLockScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

// AdapterRedis is the mcache adapter implements using Redis server.
type AdapterRedis struct {
	redis             *mredis.Redis
	lockTTL           time.Duration
	lockRetryInterval time.Duration
	lockWaitTimeout   time.Duration
}

// Option configures a Redis cache adapter.
type Option func(*AdapterRedis)

// WithLockTTL sets how long a distributed cache lock remains valid.
func WithLockTTL(ttl time.Duration) Option {
	return func(adapter *AdapterRedis) {
		if ttl > 0 {
			adapter.lockTTL = ttl
		}
	}
}

// WithLockRetryInterval sets how often a waiter retries an occupied lock.
func WithLockRetryInterval(interval time.Duration) Option {
	return func(adapter *AdapterRedis) {
		if interval > 0 {
			adapter.lockRetryInterval = interval
		}
	}
}

// WithLockWaitTimeout limits how long GetOrSetFuncLock waits for a value or lock.
func WithLockWaitTimeout(timeout time.Duration) Option {
	return func(adapter *AdapterRedis) {
		if timeout > 0 {
			adapter.lockWaitTimeout = timeout
		}
	}
}

// NewAdapterRedis creates and returns a new redis adapter for mcache.
func NewAdapterRedis(redisClient *mredis.Redis, options ...Option) mcache.Adapter {
	adapter := &AdapterRedis{
		redis:             redisClient,
		lockTTL:           defaultLockTTL,
		lockRetryInterval: defaultLockRetryInterval,
		lockWaitTimeout:   defaultLockWaitTimeout,
	}
	for _, option := range options {
		if option != nil {
			option(adapter)
		}
	}
	return adapter
}

// Set sets cache with `key`-`value` pair, which is expired after `duration`.
// It does not expire if `duration` == 0.
// It deletes the key if `duration` < 0 or given `value` is nil.
func (c *AdapterRedis) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	if value == nil || duration < 0 {
		_, err := c.redis.Del(ctx, key)
		return err
	}
	if duration == 0 {
		err := c.redis.Set(ctx, key, value)
		return err
	}
	return c.redis.SetEX(ctx, key, value, duration)
}

// Get retrieves and returns the associated value of given `key`.
// It returns nil if it does not exist, or its value is nil, or it's expired.
func (c *AdapterRedis) Get(ctx context.Context, key string) (*mvar.Var, error) {
	v, err := c.redis.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return v, nil
}

// SetMap batch sets cache with key-value pairs by `data` map, which is expired after `duration`.
func (c *AdapterRedis) SetMap(ctx context.Context, data map[string]interface{}, duration time.Duration) error {
	if len(data) == 0 {
		return nil
	}
	if duration < 0 {
		keys := make([]string, 0, len(data))
		for k := range data {
			keys = append(keys, k)
		}
		_, err := c.redis.Del(ctx, keys...)
		return err
	}
	if duration == 0 {
		return c.redis.MSet(ctx, data)
	}

	// For redis adapter, we should use pipeline to set multiple keys with expiration.
	// However, to keep it simple, we loop it here.
	// A more robust implementation might use redis.Pipelined here.
	for k, v := range data {
		if err := c.Set(ctx, k, v, duration); err != nil {
			return err
		}
	}
	return nil
}

// SetIfNotExist sets cache with `key`-`value` pair if `key` does not exist in the cache.
// It is an atomic operation.
func (c *AdapterRedis) SetIfNotExist(ctx context.Context, key string, value interface{}, duration time.Duration) (ok bool, err error) {
	if value == nil || duration < 0 {
		var n int64
		n, err = c.redis.Del(ctx, key)
		return n > 0, err
	}
	return c.redis.SetNX(ctx, key, value, duration)
}

// SetIfNotExistFunc sets `key` with result of function `f` if `key` does not exist in the cache.
func (c *AdapterRedis) SetIfNotExistFunc(ctx context.Context, key string, f mcache.Func, duration time.Duration) (ok bool, err error) {
	// Check existence first.
	v, err := c.redis.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	if v > 0 {
		return false, nil
	}
	// Execute function and set value.
	value, err := f(ctx)
	if err != nil {
		return false, err
	}
	return c.SetIfNotExist(ctx, key, value, duration)
}

// SetIfNotExistFuncLock sets `key` with result of function `f` if `key` does not exist in the cache.
// Note that the function `f` is executed within redis lock.
func (c *AdapterRedis) SetIfNotExistFuncLock(ctx context.Context, key string, f mcache.Func, duration time.Duration) (ok bool, err error) {
	lockKey := cacheLockKey(key)
	token, locked, err := c.acquireLock(ctx, lockKey)
	if err != nil {
		return false, err
	}
	if !locked {
		return false, nil
	}
	defer func() {
		if releaseErr := c.releaseLock(ctx, lockKey, token); releaseErr != nil && err == nil {
			err = releaseErr
		}
	}()

	v, err := c.redis.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	if v > 0 {
		return false, nil
	}
	value, err := f(ctx)
	if err != nil {
		return false, err
	}
	return c.SetIfNotExist(ctx, key, value, duration)
}

// GetOrSet retrieves and returns the value of `key`, or sets `key`-`value` pair and returns `value` if `key` does not exist.
func (c *AdapterRedis) GetOrSet(ctx context.Context, key string, value interface{}, duration time.Duration) (result *mvar.Var, err error) {
	result, err = c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}
	err = c.Set(ctx, key, value, duration)
	if err != nil {
		return nil, err
	}
	return mvar.New(value), nil
}

// GetOrSetFunc retrieves and returns the value of `key`, or sets `key` with result of function `f` and returns its result if `key` does not exist.
func (c *AdapterRedis) GetOrSetFunc(ctx context.Context, key string, f mcache.Func, duration time.Duration) (result *mvar.Var, err error) {
	result, err = c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}
	var value interface{}
	value, err = f(ctx)
	if err != nil {
		return nil, err
	}
	err = c.Set(ctx, key, value, duration)
	if err != nil {
		return nil, err
	}
	return mvar.New(value), nil
}

// GetOrSetFuncLock retrieves and returns the value of `key`, or sets `key` with result of function `f` and returns its result if `key` does not exist.
func (c *AdapterRedis) GetOrSetFuncLock(ctx context.Context, key string, f mcache.Func, duration time.Duration) (result *mvar.Var, err error) {
	waitCtx, cancel := context.WithTimeout(ctx, c.lockWaitTimeout)
	defer cancel()
	lockKey := cacheLockKey(key)

	for {
		result, err = c.Get(waitCtx, key)
		if err != nil {
			return nil, err
		}
		if result != nil {
			return result, nil
		}

		token, locked, lockErr := c.acquireLock(waitCtx, lockKey)
		if lockErr != nil {
			return nil, lockErr
		}
		if !locked {
			if waitErr := waitForLockRetry(waitCtx, c.lockRetryInterval); waitErr != nil {
				return nil, waitErr
			}
			continue
		}

		result, err = c.getOrSetUnderLock(waitCtx, key, f, duration)
		releaseErr := c.releaseLock(waitCtx, lockKey, token)
		if err != nil {
			return nil, err
		}
		if releaseErr != nil {
			return nil, releaseErr
		}
		return result, nil
	}
}

func (c *AdapterRedis) getOrSetUnderLock(ctx context.Context, key string, f mcache.Func, duration time.Duration) (*mvar.Var, error) {
	result, err := c.Get(ctx, key)
	if err != nil || result != nil {
		return result, err
	}

	value, err := f(ctx)
	if err != nil {
		return nil, err
	}
	if err := c.Set(ctx, key, value, duration); err != nil {
		return nil, err
	}
	return mvar.New(value), nil
}

func (c *AdapterRedis) acquireLock(ctx context.Context, lockKey string) (token string, locked bool, err error) {
	token, err = newLockToken()
	if err != nil {
		return "", false, err
	}
	locked, err = c.redis.SetNX(ctx, lockKey, token, c.lockTTL)
	return token, locked, err
}

func (c *AdapterRedis) releaseLock(ctx context.Context, lockKey, token string) error {
	releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), lockReleaseTimeout)
	defer cancel()
	return c.redis.Client().Eval(releaseCtx, releaseLockScript, []string{lockKey}, token).Err()
}

func newLockToken() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(token[:]), nil
}

func cacheLockKey(key string) string {
	return key + lockKeySuffix
}

func waitForLockRetry(ctx context.Context, interval time.Duration) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Contains checks and returns true if `key` exists in the cache, or else returns false.
func (c *AdapterRedis) Contains(ctx context.Context, key string) (bool, error) {
	v, err := c.redis.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return v > 0, nil
}

// Size returns the number of items in the cache.
// Note that this method is not accurate in redis cluster mode and may be slow if the cache size is large.
func (c *AdapterRedis) Size(ctx context.Context) (size int, err error) {
	n, err := c.redis.DBSize(ctx)
	return int(n), err
}

// Data returns a copy of all key-value pairs in the selected Redis database.
// It incrementally scans keys, but may still be expensive for a large database.
func (c *AdapterRedis) Data(ctx context.Context) (data map[string]interface{}, err error) {
	keys, err := c.scanKeys(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}
	values, err := c.redis.MGet(ctx, keys...)
	if err != nil {
		return nil, err
	}
	data = make(map[string]interface{}, len(keys))
	for i, k := range keys {
		if i < len(values) {
			data[k] = values[i].Val()
		}
	}
	return data, nil
}

// Keys returns all keys in the selected Redis database using incremental SCAN calls.
func (c *AdapterRedis) Keys(ctx context.Context) (keys []string, err error) {
	return c.scanKeys(ctx)
}

// Values returns all values in the selected Redis database.
// It incrementally scans keys, but may still be expensive for a large database.
func (c *AdapterRedis) Values(ctx context.Context) (values []interface{}, err error) {
	keys, err := c.scanKeys(ctx)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return make([]interface{}, 0), nil
	}
	vars, err := c.redis.MGet(ctx, keys...)
	if err != nil {
		return nil, err
	}
	values = make([]interface{}, len(vars))
	for i, v := range vars {
		values[i] = v.Val()
	}
	return values, nil
}

func (c *AdapterRedis) scanKeys(ctx context.Context) ([]string, error) {
	iterator := c.redis.Client().Scan(ctx, 0, "*", 100).Iterator()
	keys := make([]string, 0)
	for iterator.Next(ctx) {
		keys = append(keys, iterator.Val())
	}
	if err := iterator.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

// Update updates the value of `key` without changing its expiration and returns the old value.
func (c *AdapterRedis) Update(ctx context.Context, key string, value interface{}) (oldValue *mvar.Var, exist bool, err error) {
	oldValue, err = c.Get(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if oldValue == nil {
		return nil, false, nil
	}
	ttl, err := c.redis.TTL(ctx, key)
	if err != nil {
		return nil, true, err
	}
	err = c.Set(ctx, key, value, ttl)
	if err != nil {
		return nil, true, err
	}
	return oldValue, true, nil
}

// UpdateExpire updates the expiration of `key` and returns the old expiration duration value.
func (c *AdapterRedis) UpdateExpire(ctx context.Context, key string, duration time.Duration) (oldDuration time.Duration, err error) {
	ttl, err := c.redis.TTL(ctx, key)
	if err != nil {
		return -1, err
	}
	if ttl <= -2 { // Key does not exist
		return -1, nil
	}
	if ttl == -1 { // Key has no expiration.
		oldDuration = 0
	} else {
		oldDuration = ttl
	}

	if duration < 0 {
		_, err = c.redis.Del(ctx, key)
		return
	}
	if duration == 0 {
		_, err = c.redis.Persist(ctx, key)
		return
	}
	_, err = c.redis.Expire(ctx, key, duration)
	return
}

// GetExpire retrieves and returns the expiration of `key` in the cache.
func (c *AdapterRedis) GetExpire(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.redis.TTL(ctx, key)
	if err != nil {
		return 0, err
	}
	if ttl == -2 { // Key does not exist.
		return -1, nil
	}
	if ttl == -1 { // Key has no expiration.
		return 0, nil
	}
	return time.Duration(ttl) * time.Second, nil
}

// Remove deletes one or more keys from cache.
func (c *AdapterRedis) Remove(ctx context.Context, keys ...string) (lastValue *mvar.Var, err error) {
	if len(keys) == 0 {
		return nil, nil
	}
	lastValue, err = c.Get(ctx, keys[len(keys)-1])
	if err != nil {
		return nil, err
	}
	_, err = c.redis.Del(ctx, keys...)
	return
}

// Clear clears all data of the cache.
func (c *AdapterRedis) Clear(ctx context.Context) error {
	return c.redis.FlushDB(ctx)
}

// Close closes the cache.
func (c *AdapterRedis) Close(_ context.Context) error {
	// A redis adapter should not close the redis client,
	// as the client might be shared.
	return nil
}
