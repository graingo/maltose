package mredis

import (
	"context"

	"github.com/graingo/maltose/container/mvar"
	"github.com/redis/go-redis/v9"
)

// LPush prepends one or multiple values to a list.
func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return r.client.LPush(ctx, key, values...).Result()
}

// RPop removes and returns the last element of the list stored at key.
func (r *Redis) RPop(ctx context.Context, key string) (*mvar.Var, error) {
	val, err := r.client.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return mvar.New(val), nil
}
