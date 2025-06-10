package mredis

import "context"

// Del removes the specified keys. A key is ignored if it does not exist.
func (r *Redis) Del(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Del(ctx, keys...).Result()
}

// Exists returns if key exists.
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}
