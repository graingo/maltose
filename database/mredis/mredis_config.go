package mredis

import (
	"time"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/os/mlog"
	"github.com/graingo/mconv"
)

// Config is the configuration object for Redis.
// Zero-value fields inherit the framework defaults when passed to New.
type Config struct {
	// Address is the address of the Redis server.
	Address string `mconv:"address"`
	// DB is the database number.
	DB int `mconv:"db"`
	// User is the user of the Redis server.
	User string `mconv:"user"`
	// Password is the password of the Redis server.
	Password string `mconv:"password"`
	// MasterName is the master name of the Redis server.
	MasterName string `mconv:"master_name"`
	// MinIdleConns is the minimum number of idle connections.
	MinIdleConns int `mconv:"min_idle_conns"`
	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int `mconv:"max_idle_conns"`
	// MaxRetries is the maximum number of retries before giving up.
	MaxRetries int `mconv:"max_retries"`
	// PoolSize is the maximum number of socket connections.
	PoolSize int `mconv:"pool_size"`
	// MinRetryBackoff is the minimum backoff between each retry.
	MinRetryBackoff time.Duration `mconv:"min_retry_backoff"`
	// MaxRetryBackoff is the maximum backoff between each retry.
	MaxRetryBackoff time.Duration `mconv:"max_retry_backoff"`
	// DialTimeout is the timeout for establishing new connections.
	DialTimeout time.Duration `mconv:"dial_timeout"`
	// ReadTimeout is the timeout for reading.
	ReadTimeout time.Duration `mconv:"read_timeout"`
	// WriteTimeout is the timeout for writing.
	WriteTimeout time.Duration `mconv:"write_timeout"`
	// PoolTimeout is the timeout for getting a connection from the pool.
	PoolTimeout time.Duration `mconv:"pool_timeout"`
	// ConnMaxIdleTime is the timeout for idle connections.
	ConnMaxIdleTime time.Duration `mconv:"conn_max_idle_time"`
	// SlowThreshold is the slow threshold for the Redis.
	SlowThreshold time.Duration `mconv:"slow_threshold"`
	// Logger is the logger for the Redis.
	Logger *mlog.Logger
	// Hooks is the hooks for the Redis. It will be used to add hooks to the Redis client.
	Hooks []Hook
	// loggerHook is the internal logger hook instance.
	loggerHook Hook `mconv:"-"`
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

func cloneConfig(config *Config) *Config {
	if config == nil {
		return defaultConfig()
	}
	cloned := *config
	cloned.Hooks = append([]Hook(nil), config.Hooks...)
	cloned.loggerHook = nil
	return &cloned
}

// mergeConfig overlays non-zero user values on the Redis defaults.
func mergeConfig(config *Config) *Config {
	merged := defaultConfig()
	if config == nil {
		return merged
	}

	if config.Address != "" {
		merged.Address = config.Address
	}
	merged.DB = config.DB
	if config.User != "" {
		merged.User = config.User
	}
	if config.Password != "" {
		merged.Password = config.Password
	}
	if config.MasterName != "" {
		merged.MasterName = config.MasterName
	}
	if config.MinIdleConns != 0 {
		merged.MinIdleConns = config.MinIdleConns
	}
	if config.MaxIdleConns != 0 {
		merged.MaxIdleConns = config.MaxIdleConns
	}
	if config.MaxRetries != 0 {
		merged.MaxRetries = config.MaxRetries
	}
	if config.PoolSize != 0 {
		merged.PoolSize = config.PoolSize
	}
	if config.MinRetryBackoff != 0 {
		merged.MinRetryBackoff = config.MinRetryBackoff
	}
	if config.MaxRetryBackoff != 0 {
		merged.MaxRetryBackoff = config.MaxRetryBackoff
	}
	if config.DialTimeout != 0 {
		merged.DialTimeout = config.DialTimeout
	}
	if config.ReadTimeout != 0 {
		merged.ReadTimeout = config.ReadTimeout
	}
	if config.WriteTimeout != 0 {
		merged.WriteTimeout = config.WriteTimeout
	}
	if config.PoolTimeout != 0 {
		merged.PoolTimeout = config.PoolTimeout
	}
	if config.ConnMaxIdleTime != 0 {
		merged.ConnMaxIdleTime = config.ConnMaxIdleTime
	}
	if config.SlowThreshold != 0 {
		merged.SlowThreshold = config.SlowThreshold
	}
	if config.Logger != nil {
		merged.Logger = config.Logger
	}
	if config.Hooks != nil {
		merged.Hooks = append([]Hook(nil), config.Hooks...)
	}

	return merged
}

func (c *Config) SetConfigWithMap(config map[string]any) error {
	return mconv.ToStructE(config, c)
}

func (c *Config) SetLogger(logger *mlog.Logger) {
	if logger == nil {
		logger = mlog.New()
	}
	c.Logger = logger.With(mlog.String(maltose.COMPONENT, "mredis"))
}

func (c *Config) AddHook(hook Hook) {
	c.Hooks = append(c.Hooks, hook)
}
