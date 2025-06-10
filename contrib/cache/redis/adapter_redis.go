package redis

import (
	"context"
	"time"

	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/database/mredis"
	"github.com/graingo/maltose/os/mcache"
)

// AdapterRedis is the mcache adapter implements using Redis server.
type AdapterRedis struct {
	redis mredis.Redis
}

// NewAdapterRedis creates and returns a new redis adapter for mcache.
func NewAdapterRedis(redis mredis.Redis) mcache.Adapter {
	return &AdapterRedis{
		redis: redis,
	}
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
		_, err := c.redis.Set(ctx, key, value)
		return err
	}
	_, err := c.redis.SetEX(ctx, key, value, int64(duration.Seconds()))
	return err
}

// Get retrieves and returns the associated value of given `key`.
// It returns nil if it does not exist, or its value is nil, or it's expired.
func (c *AdapterRedis) Get(ctx context.Context, key string) (*mvar.Var, error) {
	return c.redis.Get(ctx, key)
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
func (c *AdapterRedis) SetIfNotExist(ctx context.Context, key string, value interface{}, duration time.Duration) (ok bool, err error) {
	if value == nil || duration < 0 {
		var n int64
		n, err = c.redis.Del(ctx, key)
		return n > 0, err
	}
	ok, err = c.redis.SetNX(ctx, key, value)
	if err != nil || !ok {
		return ok, err
	}
	if duration > 0 {
		_, err = c.redis.Expire(ctx, key, int64(duration.Seconds()))
	}
	return
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
	// Use redis lock to ensure atomicity.
	// This is a simplified lock, a more robust implementation should use a distributed lock library.
	lockKey := key + "_lock"
	locked, err := c.redis.SetNX(ctx, lockKey, 1)
	if err != nil {
		return false, err
	}
	if !locked {
		return false, nil // Another process is setting the value.
	}
	defer c.redis.Del(ctx, lockKey)

	// Double check inside the lock.
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
	if result.IsNil() {
		err = c.Set(ctx, key, value, duration)
		if err != nil {
			return nil, err
		}
		return mvar.New(value), nil
	}
	return result, nil
}

// GetOrSetFunc retrieves and returns the value of `key`, or sets `key` with result of function `f` and returns its result if `key` does not exist.
func (c *AdapterRedis) GetOrSetFunc(ctx context.Context, key string, f mcache.Func, duration time.Duration) (result *mvar.Var, err error) {
	result, err = c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if result.IsNil() {
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
	return result, nil
}

// GetOrSetFuncLock retrieves and returns the value of `key`, or sets `key` with result of function `f` and returns its result if `key` does not exist.
func (c *AdapterRedis) GetOrSetFuncLock(ctx context.Context, key string, f mcache.Func, duration time.Duration) (result *mvar.Var, err error) {
	result, err = c.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if result.IsNil() {
		// Use redis lock to ensure atomicity.
		lockKey := key + "_lock"
		locked, err := c.redis.SetNX(ctx, lockKey, 1)
		if err != nil {
			return nil, err
		}
		if locked {
			defer c.redis.Del(ctx, lockKey)
			// Double check inside the lock.
			result, err = c.Get(ctx, key)
			if err != nil {
				return nil, err
			}
			if result.IsNil() {
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
		} else {
			// Wait and retry getting the value.
			time.Sleep(50 * time.Millisecond)
			return c.Get(ctx, key)
		}
	}
	return result, nil
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
func (c *AdapterRedis) Size(ctx context.Context) (size int, err error) {
	n, err := c.redis.DBSize(ctx)
	return int(n), err
}

// Data returns a copy of all key-value pairs in the cache as map type.
func (c *AdapterRedis) Data(ctx context.Context) (data map[string]interface{}, err error) {
	keys, err := c.redis.Keys(ctx, "*")
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

// Keys returns all keys in the cache as slice.
func (c *AdapterRedis) Keys(ctx context.Context) (keys []string, err error) {
	return c.redis.Keys(ctx, "*")
}

// Values returns all values in the cache as slice.
func (c *AdapterRedis) Values(ctx context.Context) (values []interface{}, err error) {
	keys, err := c.redis.Keys(ctx, "*")
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

// Update updates the value of `key` without changing its expiration and returns the old value.
func (c *AdapterRedis) Update(ctx context.Context, key string, value interface{}) (oldValue *mvar.Var, exist bool, err error) {
	oldValue, err = c.Get(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if oldValue.IsNil() {
		return nil, false, nil
	}
	ttl, err := c.redis.TTL(ctx, key)
	if err != nil {
		return nil, false, err
	}
	err = c.Set(ctx, key, value, time.Duration(ttl)*time.Second)
	if err != nil {
		return nil, true, err
	}
	return oldValue, true, nil
}

// UpdateExpire updates the expiration of `key` and returns the old expiration duration value.
func (c *AdapterRedis) UpdateExpire(ctx context.Context, key string, duration time.Duration) (oldDuration time.Duration, err error) {
	ttl, err := c.redis.TTL(ctx, key)
	if err != nil || ttl <= -2 { // Key does not exist
		return -1, err
	}
	oldDuration = time.Duration(ttl) * time.Second
	if duration < 0 {
		_, err = c.redis.Del(ctx, key)
		return
	}
	_, err = c.redis.Expire(ctx, key, int64(duration.Seconds()))
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
func (c *AdapterRedis) Close(ctx context.Context) error {
	// A redis adapter should not close the redis client,
	// as the client might be shared.
	return nil
}
