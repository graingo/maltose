package mredis

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/spf13/viper"
)

// Config is the configuration object for Redis.
type Config struct {
	// Address is the address of the Redis server.
	Address string `mapstructure:"address"`
	// DB is the database number.
	DB int `mapstructure:"db"`
	// User is the user of the Redis server.
	User string `mapstructure:"user"`
	// Password is the password of the Redis server.
	Password string `mapstructure:"password"`
	// MasterName is the master name of the Redis server.
	MasterName string `mapstructure:"master_name"`
	// MinIdleConns is the minimum number of idle connections.
	MinIdleConns int `mapstructure:"min_idle_conns"`
	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int `mapstructure:"max_idle_conns"`
	// MaxRetries is the maximum number of retries before giving up.
	MaxRetries int `mapstructure:"max_retries"`
	// PoolSize is the maximum number of socket connections.
	PoolSize int `mapstructure:"pool_size"`
	// MinRetryBackoff is the minimum backoff between each retry.
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff"`
	// MaxRetryBackoff is the maximum backoff between each retry.
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff"`
	// DialTimeout is the timeout for establishing new connections.
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
	// ReadTimeout is the timeout for reading.
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	// WriteTimeout is the timeout for writing.
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	// PoolTimeout is the timeout for getting a connection from the pool.
	PoolTimeout time.Duration `mapstructure:"pool_timeout"`
	// ConnMaxIdleTime is the timeout for idle connections.
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	// SlowThreshold is the slow threshold for the Redis.
	SlowThreshold time.Duration `mapstructure:"slow_threshold"`
	// Logger is the logger for the Redis.
	Logger *mlog.Logger
	// Hooks is the hooks for the Redis. It will be used to add hooks to the Redis client.
	Hooks []Hook
}

func defaultConfig() *Config {
	return &Config{
		Address:      "127.0.0.1:6379",
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		DialTimeout:  5 * time.Second,
		PoolSize:     10,
	}
}

func (c *Config) SetConfigWithMap(config map[string]any) error {
	v := viper.New()
	v.MergeConfigMap(config)
	return v.Unmarshal(c)
}

func (c *Config) SetLogger(logger *mlog.Logger) {
	c.Logger = logger
}

func (c *Config) AddHook(hook Hook) {
	c.Hooks = append(c.Hooks, hook)
}
