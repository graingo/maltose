package mdb

import (
	"time"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/os/mlog"
	"github.com/graingo/mconv"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
)

type Config struct {
	// Type is the type of the database.
	Type string `mconv:"type"`
	// DSN is the data source name.
	DSN string `mconv:"dsn"`
	// Host is the host of the database.
	Host string `mconv:"host"`
	// Port is the port of the database.
	Port string `mconv:"port"`
	// User is the user of the database.
	User string `mconv:"user"`
	// Password is the password of the database.
	Password string `mconv:"password"`
	// DBName is the name of the database.
	DBName string `mconv:"db_name"`
	// MaxIdleTime is the maximum idle time for the database connection.
	MaxIdleTime time.Duration `mconv:"max_idle_time"`
	// MaxIdleConnection is the maximum idle connection for the database.
	MaxIdleConnection int `mconv:"max_idle_connection"`
	// MaxOpenConnection is the maximum open connection for the database.
	MaxOpenConnection int `mconv:"max_open_connection"`
	// MaxLifetime is the maximum lifetime for the database connection.
	MaxLifetime time.Duration `mconv:"max_lifetime"`
	// SlowThreshold is the slow query threshold.
	SlowThreshold time.Duration `mconv:"slow_threshold"`
	// Logger is the logger for the database.
	Logger *mlog.Logger
	// Replicas is the replicas list.
	Replicas []Config
	// Plugins is the plugins list.
	Plugins []gorm.Plugin
}

func defaultConfig() *Config {
	return &Config{
		Type:              "mysql",
		Port:              "3306",
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
	return mconv.ToStructE(config, c)
}

func (c *Config) SetLogger(logger *mlog.Logger) {
	if logger == nil {
		// If no logger is provided, create a default one.
		// The component field will be added by the GormLogger wrapper.
		logger = mlog.New()
	}
	c.Logger = logger.With(mlog.String(maltose.COMPONENT, "mdb"))
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
