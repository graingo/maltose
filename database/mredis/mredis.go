package mredis

import (
	"context"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// Redis is the main struct for redis operations.
type Redis struct {
	client redis.UniversalClient
	config *Config
}

type Hook redis.Hook

// New creates and returns a new Redis client.
func New(config ...*Config) (*Redis, error) {
	cfg := defaultConfig()
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}
	if cfg == nil {
		return nil, merror.NewCode(
			mcode.CodeInvalidConfiguration,
			`no configuration found for creating Redis client`,
		)
	}

	// more options can be added here
	opts := &redis.UniversalOptions{
		Addrs:           []string{cfg.Address},
		DB:              cfg.DB,
		Username:        cfg.User,
		Password:        cfg.Password,
		MasterName:      cfg.MasterName,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxRetries:      cfg.MaxRetries,
		PoolSize:        cfg.PoolSize,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}

	client := redis.NewUniversalClient(opts)

	// Add logger hook if logger is configured.
	if cfg.Logger != nil {
		client.AddHook(newLoggerHook(cfg.Logger, cfg.SlowThreshold))
	}

	// Enable tracing.
	if err := redisotel.InstrumentTracing(client); err != nil {
		panic(err)
	}

	// Add hooks from config
	for _, hook := range cfg.Hooks {
		client.AddHook(hook)
	}

	return &Redis{
		client: client,
		config: cfg,
	}, nil
}

// Client returns the underlying universal client.
func (r *Redis) Client() redis.UniversalClient {
	return r.client
}

// AddHook adds a hook to the client.
func (r *Redis) AddHook(hook Hook) {
	r.config.Hooks = append(r.config.Hooks, hook)
	r.client.AddHook(hook)
}

// Ping checks the connection to the server.
func (r *Redis) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the client, releasing any open resources.
func (r *Redis) Close() error {
	return r.client.Close()
}
