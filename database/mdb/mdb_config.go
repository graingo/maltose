package mdb

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/spf13/viper"
)

type Config struct {
	Type              string
	DSN               string
	Host              string
	Port              string
	User              string
	Password          string
	DBName            string
	MaxIdleTime       time.Duration
	MaxIdleConnection int
	MaxOpenConnection int
	Logger            *mlog.Logger
	SlowThreshold     time.Duration // slow query threshold
}

func DefaultConfig() *Config {
	return &Config{
		Type:              "mysql",
		MaxIdleTime:       10 * time.Second,
		MaxIdleConnection: 10,
		MaxOpenConnection: 100,
		Logger:            mlog.New(),
		SlowThreshold:     300 * time.Millisecond,
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
