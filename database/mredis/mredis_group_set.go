package mredis

import (
	"context"
)

// SAdd adds the specified members to the set stored at key.
func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(ctx, key, members...).Result()
}

// SIsMember returns if member is a member of the set stored at key.
func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.client.SIsMember(ctx, key, member).Result()
}
