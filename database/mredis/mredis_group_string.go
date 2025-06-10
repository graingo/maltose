package mredis

import (
	"context"
	"time"

	"github.com/graingo/maltose/container/mvar"
)

// Set sets key to hold the string value.
// If key already holds a value, it is overwritten, regardless of its type.
// Any previous time to live associated with the key is discarded on successful SET operation.
func (r *Redis) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	return r.client.Set(ctx, key, value, duration).Err()
}

// Get gets the value of key.
// If the key does not exist the special value nil is returned.
// An error is returned if the value stored at key is not a string, because GET only handles string values.
func (r *Redis) Get(ctx context.Context, key string) (*mvar.Var, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return mvar.New(val), nil
}
