package mredis

import (
	"time"

	"github.com/spf13/viper"
)

// Config is the configuration for the Redis client.
type Config struct {
	// Address can be a single address or a comma-separated list of addresses for a cluster.
	Address string `json:"address"`
	// DB is the database to select after connecting to the server.
	DB int `json:"db"`
	// User is the user to authenticate with.
	User string `json:"user"`
	// Password is the password to authenticate with.
	Password string `json:"password"`
	// MasterName is the name of the master for Sentinel connections.
	MasterName string `json:"masterName"`
	// MinIdleConns is the minimum number of idle connections.
	MinIdleConns int `json:"minIdleConns"`
	// MaxIdleConns is the maximum number of idle connections.
	MaxIdleConns int `json:"maxIdleConns"`
	// MaxRetries is the maximum number of retries before giving up.
	MaxRetries int `json:"maxRetries"`
	// PoolSize is the maximum number of socket connections.
	PoolSize int `json:"poolSize"`
	// MinRetryBackoff is the minimum backoff between each retry.
	MinRetryBackoff time.Duration `json:"minRetryBackoff"`
	// MaxRetryBackoff is the maximum backoff between each retry.
	MaxRetryBackoff time.Duration `json:"maxRetryBackoff"`
	// DialTimeout is the timeout for establishing new connections.
	DialTimeout time.Duration `json:"dialTimeout"`
	// ReadTimeout is the timeout for reading.
	ReadTimeout time.Duration `json:"readTimeout"`
	// WriteTimeout is the timeout for writing.
	WriteTimeout time.Duration `json:"writeTimeout"`
	// PoolTimeout is the timeout for getting a connection from the pool.
	PoolTimeout time.Duration `json:"poolTimeout"`
	// ConnMaxIdleTime is the timeout for idle connections.
	ConnMaxIdleTime time.Duration `json:"idleTimeout"`
}

func (c *Config) SetConfigWithMap(config map[string]any) error {
	v := viper.New()
	v.MergeConfigMap(config)
	return v.Unmarshal(c)
}
