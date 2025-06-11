package mredis

import (
	"context"

	"github.com/graingo/maltose/container/mvar"
	"github.com/redis/go-redis/v9"
)

// HSet sets the specified fields to their respective values in the hash stored at key.
func (r *Redis) HSet(ctx context.Context, key string, fields map[string]interface{}) error {
	return r.client.HSet(ctx, key, fields).Err()
}

// HGet returns the value associated with field in the hash stored at key.
func (r *Redis) HGet(ctx context.Context, key, field string) (*mvar.Var, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return mvar.New(val), nil
}
