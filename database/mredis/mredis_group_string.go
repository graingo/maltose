package mredis

import (
	"context"
	"time"

	"github.com/graingo/maltose/container/mvar"
)

// Set sets key to hold the string value.
// If key already holds a value, it is overwritten, regardless of its type.
// Any previous time to live associated with the key is discarded on successful SET operation.
func (r *Redis) Set(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

// SetEx sets key to hold the string value with time.
func (r *Redis) SetEX(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	return r.client.SetEx(ctx, key, value, duration).Err()
}

// Get retrieves and returns the associated value of given `key`.
// If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string, because GET only handles string values.
func (r *Redis) Get(ctx context.Context, key string) (*mvar.Var, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return mvar.New(val), nil
}

// MSet is a redis MSET command.
func (r *Redis) MSet(ctx context.Context, data map[string]interface{}) error {
	return r.client.MSet(ctx, data).Err()
}

// MGet is a redis MGET command.
func (r *Redis) MGet(ctx context.Context, keys ...string) ([]*mvar.Var, error) {
	result, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	vars := make([]*mvar.Var, len(result))
	for i, v := range result {
		vars[i] = mvar.New(v)
	}
	return vars, nil
}

// SetNX is a redis SETNX command.
func (r *Redis) SetNX(ctx context.Context, key string, value interface{}) (bool, error) {
	return r.client.SetNX(ctx, key, value, 0).Result()
}
