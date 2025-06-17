package mredis

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/spf13/viper"
)

// Config is the configuration for the Redis client.
type Config struct {
	Address         string        `json:"address"`
	DB              int           `json:"db"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	MasterName      string        `json:"masterName"`      // sentinel master name
	MinIdleConns    int           `json:"minIdleConns"`    // minimum number of idle connections
	MaxIdleConns    int           `json:"maxIdleConns"`    // maximum number of idle connections
	MaxRetries      int           `json:"maxRetries"`      // maximum number of retries before giving up
	PoolSize        int           `json:"poolSize"`        // maximum number of socket connections
	MinRetryBackoff time.Duration `json:"minRetryBackoff"` // minimum backoff between each retry
	MaxRetryBackoff time.Duration `json:"maxRetryBackoff"` // maximum backoff between each retry
	DialTimeout     time.Duration `json:"dialTimeout"`     // timeout for establishing new connections
	ReadTimeout     time.Duration `json:"readTimeout"`     // timeout for reading
	WriteTimeout    time.Duration `json:"writeTimeout"`    // timeout for writing
	PoolTimeout     time.Duration `json:"poolTimeout"`     // timeout for getting a connection from the pool
	ConnMaxIdleTime time.Duration `json:"connMaxIdleTime"` // timeout for idle connections
	SlowThreshold   time.Duration `json:"slowThreshold"`   // slow query threshold
	Logger          *mlog.Logger  `json:"-"`
	Hooks           []Hook        `json:"-"`
}

func (c *Config) SetConfigWithMap(config map[string]any) error {
	v := viper.New()
	v.MergeConfigMap(config)
	return v.Unmarshal(c)
}

func (c *Config) SetLogger(logger *mlog.Logger) {
	c.Logger = logger
}
