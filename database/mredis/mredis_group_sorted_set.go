package mredis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// ZAdd adds all the specified members with the specified scores to the sorted set stored at key.
func (r *Redis) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return r.client.ZAdd(ctx, key, members...).Result()
}

// ZScore returns the score of member in the sorted set at key.
func (r *Redis) ZScore(ctx context.Context, key, member string) (float64, error) {
	return r.client.ZScore(ctx, key, member).Result()
}
