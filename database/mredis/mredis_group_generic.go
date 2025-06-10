package mredis

import (
	"context"
	"time"
)

// Del is a redis DEL command.
func (r *Redis) Del(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Del(ctx, keys...).Result()
}

// Exists is a redis EXISTS command.
func (r *Redis) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

// Expire is a redis EXPIRE command.
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, expiration).Result()
}

// Keys is a redis KEYS command.
func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// TTL is a redis TTL command.
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// DBSize is a redis DBSIZE command.
func (r *Redis) DBSize(ctx context.Context) (int64, error) {
	return r.client.DBSize(ctx).Result()
}

// FlushDB is a redis FLUSHDB command.
func (r *Redis) FlushDB(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}
