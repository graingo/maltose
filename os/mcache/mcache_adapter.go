package mcache

import (
	"context"
	"time"

	"github.com/graingo/maltose/container/mvar"
)

// Func is the cache function that calculates and returns the value.
type Func func(ctx context.Context) (value interface{}, err error)

// Adapter is the adapter for cache features.
type Adapter interface {
	// Set sets cache with `key`-`value` pair, which is expired after `duration`.
	// It does not expire if `duration` is 0.
	// It deletes the key if `duration` < 0 or given `value` is nil.
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) error

	// SetMap batch sets cache with key-value pairs by `data` map, which is expired after `duration`.
	SetMap(ctx context.Context, data map[string]interface{}, duration time.Duration) error

	// SetIfNotExist sets cache with `key`-`value` pair which is expired after `duration`
	// if `key` does not exist in the cache. It returns true if the `key` does not exist in the
	// cache and it sets `value` successfully to the cache, otherwise it returns false.
	SetIfNotExist(ctx context.Context, key string, value interface{}, duration time.Duration) (ok bool, err error)

	// SetIfNotExistFunc sets `key` with result of function `f` and returns true if `key` does not exist in the cache,
	// or else it does nothing and returns false if `key` already exists.
	SetIfNotExistFunc(ctx context.Context, key string, f Func, duration time.Duration) (ok bool, err error)

	// SetIfNotExistFuncLock sets `key` with result of function `f` and returns true if `key` does not exist in the cache,
	// or else it does nothing and returns false if `key` already exists.
	// It executes function `f` within writing mutex lock for concurrent safety purpose.
	SetIfNotExistFuncLock(ctx context.Context, key string, f Func, duration time.Duration) (ok bool, err error)

	// Get retrieves and returns the associated value of given `key`.
	// It returns nil if it does not exist, or its value is nil, or it's expired.
	Get(ctx context.Context, key string) (*mvar.Var, error)

	// GetOrSet retrieves and returns the value of `key`, or sets `key`-`value` pair and
	// returns `value` if `key` does not exist in the cache. The key-value pair expires after `duration`.
	GetOrSet(ctx context.Context, key string, value interface{}, duration time.Duration) (result *mvar.Var, err error)

	// GetOrSetFunc retrieves and returns the value of `key`, or sets `key` with result of
	// function `f` and returns its result if `key` does not exist in the cache.
	GetOrSetFunc(ctx context.Context, key string, f Func, duration time.Duration) (result *mvar.Var, err error)

	// GetOrSetFuncLock retrieves and returns the value of `key`, or sets `key` with result of
	// function `f` and returns its result if `key` does not exist in the cache.
	// It executes function `f` within writing mutex lock for concurrent safety purpose.
	GetOrSetFuncLock(ctx context.Context, key string, f Func, duration time.Duration) (result *mvar.Var, err error)

	// Contains checks and returns true if `key` exists in the cache, or else returns false.
	Contains(ctx context.Context, key string) (bool, error)

	// Size returns the number of items in the cache.
	Size(ctx context.Context) (size int, err error)

	// Data returns a copy of all key-value pairs in the cache as map type.
	Data(ctx context.Context) (data map[string]interface{}, err error)

	// Keys returns all keys in the cache as slice.
	Keys(ctx context.Context) (keys []string, err error)

	// Values returns all values in the cache as slice.
	Values(ctx context.Context) (values []interface{}, err error)

	// Update updates the value of `key` without changing its expiration and returns the old value.
	Update(ctx context.Context, key string, value interface{}) (oldValue *mvar.Var, exist bool, err error)

	// UpdateExpire updates the expiration of `key` and returns the old expiration duration value.
	UpdateExpire(ctx context.Context, key string, duration time.Duration) (oldDuration time.Duration, err error)

	// GetExpire retrieves and returns the expiration of `key` in the cache.
	GetExpire(ctx context.Context, key string) (time.Duration, error)

	// Remove deletes one or more keys from cache.
	Remove(ctx context.Context, keys ...string) (lastValue *mvar.Var, err error)

	// Clear clears all data of the cache.
	Clear(ctx context.Context) error

	// Close closes the cache if necessary.
	Close(ctx context.Context) error
}
