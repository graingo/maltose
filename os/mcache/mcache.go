package mcache

import (
	"context"
	"time"

	"github.com/graingo/maltose/container/mvar"
)

var (
	// defaultCache is the default cache for package method usage.
	defaultCache = New()
)

// Set sets cache with `key`-`value` pair, which is expired after `duration`.
// It does not expire if `duration` == 0.
func Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	return defaultCache.adapter.Set(ctx, key, value, duration)
}

// SetMap batch sets cache with key-value pairs by `data` map, which is expired after `duration`.
// It does not expire if `duration` == 0.
func SetMap(ctx context.Context, data map[string]interface{}, duration time.Duration) error {
	return defaultCache.adapter.SetMap(ctx, data, duration)
}

// SetIfNotExist sets cache with `key`-`value` pair if `key` does not exist in the cache.
// It returns true if `key` is set, or returns false if the `key` already exists.
func SetIfNotExist(ctx context.Context, key string, value interface{}, duration time.Duration) (bool, error) {
	return defaultCache.adapter.SetIfNotExist(ctx, key, value, duration)
}

// SetIfNotExistFunc sets `key` with result of function `f` if `key` does not exist in the cache.
// It returns true if `key` is set, or returns false if the `key` already exists.
// The function `f` is executed only if `key` does not exist in the cache.
func SetIfNotExistFunc(ctx context.Context, key string, f Func, duration time.Duration) (bool, error) {
	return defaultCache.adapter.SetIfNotExistFunc(ctx, key, f, duration)
}

// SetIfNotExistFuncLock sets `key` with result of function `f` if `key` does not exist in the cache.
// It returns true if `key` is set, or returns false if the `key` already exists.
// The function `f` is executed only if `key` does not exist in the cache.
// It is recommended to use this function instead of `SetIfNotExistFunc` if you think there might be concurrent insertions to the same `key`.
func SetIfNotExistFuncLock(ctx context.Context, key string, f Func, duration time.Duration) (bool, error) {
	return defaultCache.adapter.SetIfNotExistFuncLock(ctx, key, f, duration)
}

// Get retrieves and returns the associated value of given `key`.
// It returns nil if it does not exist or its value is nil.
func Get(ctx context.Context, key string) (*mvar.Var, error) {
	return defaultCache.adapter.Get(ctx, key)
}

// GetOrSet retrieves and returns the value of `key`, or sets `key`-`value` pair and
// returns `value` if `key` does not exist in the cache.
// The key-value pair expires after `duration`.
// It does not expire if `duration` == 0.
func GetOrSet(ctx context.Context, key string, value interface{}, duration time.Duration) (*mvar.Var, error) {
	return defaultCache.adapter.GetOrSet(ctx, key, value, duration)
}

// GetOrSetFunc retrieves and returns the value of `key`, or sets `key` with result of
// function `f` and returns its result if `key` does not exist in the cache.
// The key-value pair expires after `duration`.
// It does not expire if `duration` == 0.
func GetOrSetFunc(ctx context.Context, key string, f Func, duration time.Duration) (*mvar.Var, error) {
	return defaultCache.adapter.GetOrSetFunc(ctx, key, f, duration)
}

// GetOrSetFuncLock retrieves and returns the value of `key`, or sets `key` with result of
// function `f` and returns its result if `key` does not exist in the cache.
// The key-value pair expires after `duration`.
// It does not expire if `duration` == 0.
// It is recommended to use this function instead of `GetOrSetFunc` if you think there might be concurrent insertions to the same `key`.
func GetOrSetFuncLock(ctx context.Context, key string, f Func, duration time.Duration) (*mvar.Var, error) {
	return defaultCache.adapter.GetOrSetFuncLock(ctx, key, f, duration)
}

// Contains checks and returns true if `key` exists in the cache, or else returns false.
func Contains(ctx context.Context, key string) (bool, error) {
	return defaultCache.adapter.Contains(ctx, key)
}

// Size returns the size of the cache.
func Size(ctx context.Context) (int, error) {
	return defaultCache.adapter.Size(ctx)
}

// Data returns a copy of all key-value pairs in the cache as map type.
func Data(ctx context.Context) (map[string]interface{}, error) {
	return defaultCache.adapter.Data(ctx)
}

// Keys returns all keys in the cache as slice.
func Keys(ctx context.Context) ([]string, error) {
	return defaultCache.adapter.Keys(ctx)
}

// Values returns all values in the cache as slice.
func Values(ctx context.Context) ([]interface{}, error) {
	return defaultCache.adapter.Values(ctx)
}

// Update updates the value of `key` without changing its expiration and returns the old value.
func Update(ctx context.Context, key string, value interface{}) (oldValue *mvar.Var, exist bool, err error) {
	return defaultCache.adapter.Update(ctx, key, value)
}

// UpdateExpire updates the expiration of `key` and returns the old expiration duration value.
func UpdateExpire(ctx context.Context, key string, duration time.Duration) (oldDuration time.Duration, err error) {
	return defaultCache.adapter.UpdateExpire(ctx, key, duration)
}

// GetExpire retrieves and returns the expiration of `key` in the cache.
func GetExpire(ctx context.Context, key string) (time.Duration, error) {
	return defaultCache.adapter.GetExpire(ctx, key)
}

// Remove deletes one or more keys from cache.
func Remove(ctx context.Context, keys ...string) (lastValue *mvar.Var, err error) {
	return defaultCache.adapter.Remove(ctx, keys...)
}

// Clear clears all data of the cache.
// Note that this function is sensitive and should be carefully used.
func Clear(ctx context.Context) error {
	return defaultCache.adapter.Clear(ctx)
}
