package config

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
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
	// IdleTimeout is the timeout for idle connections.
	IdleTimeout time.Duration `json:"idleTimeout"`
	// IdleCheckFrequency is the frequency of checking for idle connections.
	IdleCheckFrequency time.Duration `json:"idleCheckFrequency"`
}

func DefaultConfig() *Config {
	return &Config{
		Type:              "mysql",
		MaxIdleTime:       10 * time.Second,
		MaxIdleConnection: 10,
		MaxOpenConnection: 100,
		MaxLifetime:       0,
		Logger:            mlog.New(),
		SlowThreshold:     300 * time.Millisecond,
		Replicas:          []Config{},
		Plugins: []gorm.Plugin{
			otelgorm.NewPlugin(),
		},
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

func (c *Config) SetReplicas(replicas []Config) {
	c.Replicas = replicas
}

func (c *Config) AddReplica(replica Config) {
	c.Replicas = append(c.Replicas, replica)
}

func (c *Config) AddPlugin(plugins ...gorm.Plugin) {
	c.Plugins = append(c.Plugins, plugins...)
}

func (c *Config) SetPlugins(plugins []gorm.Plugin) {
	c.Plugins = plugins
}

// NewClient creates and returns a new Redis client based on the configuration.
func (c *Config) NewClient() redis.UniversalClient {
	var client redis.UniversalClient
	// more options can be added here
	opts := &redis.UniversalOptions{
		Addrs:              []string{c.Address},
		DB:                 c.DB,
		Username:           c.User,
		Password:           c.Password,
		MasterName:         c.MasterName,
		MinIdleConns:       c.MinIdleConns,
		MaxIdleConns:       c.MaxIdleConns,
		MaxRetries:         c.MaxRetries,
		PoolSize:           c.PoolSize,
		MinRetryBackoff:    c.MinRetryBackoff,
		MaxRetryBackoff:    c.MaxRetryBackoff,
		DialTimeout:        c.DialTimeout,
		ReadTimeout:        c.ReadTimeout,
		WriteTimeout:       c.WriteTimeout,
		PoolTimeout:        c.PoolTimeout,
		IdleTimeout:        c.IdleTimeout,
		IdleCheckFrequency: c.IdleCheckFrequency,
	}

	client = redis.NewUniversalClient(opts)
	return client
}
