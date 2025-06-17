package mdb

import (
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
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
	MaxLifetime       time.Duration
	Logger            *mlog.Logger
	SlowThreshold     time.Duration // slow query threshold
	Replicas          []Config      // replicas list
	Plugins           []gorm.Plugin // plugins list
}

func defaultConfig() *Config {
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
