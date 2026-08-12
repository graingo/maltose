package mdb

import (
	"context"
	"net"
	"net/url"
	"time"

	driverMySQL "github.com/go-sql-driver/mysql"
	"github.com/graingo/maltose"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mlog"
	gormMySQL "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

// CreateDriver creates a database driver based on the configuration.
func createDriver(cfg *Config) (gorm.Dialector, error) {
	// Handle DSN
	dsn := cfg.DSN
	if dsn == "" && cfg.Host != "" {
		// Build DSN from connection parameters if DSN is not provided
		switch cfg.Type {
		case "mysql":
			mysqlConfig := driverMySQL.NewConfig()
			mysqlConfig.User = cfg.User
			mysqlConfig.Passwd = cfg.Password
			mysqlConfig.Net = "tcp"
			mysqlConfig.Addr = net.JoinHostPort(cfg.Host, cfg.Port)
			mysqlConfig.DBName = cfg.DBName
			mysqlConfig.Params = map[string]string{"charset": "utf8mb4"}
			mysqlConfig.ParseTime = true
			mysqlConfig.Loc = time.Local
			dsn = mysqlConfig.FormatDSN()
		case "postgres":
			postgresURL := &url.URL{
				Scheme:   "postgres",
				User:     url.UserPassword(cfg.User, cfg.Password),
				Host:     net.JoinHostPort(cfg.Host, cfg.Port),
				Path:     cfg.DBName,
				RawQuery: "sslmode=disable",
			}
			dsn = postgresURL.String()
		}
	}

	if dsn == "" {
		return nil, merror.New("DSN is not set, please configure DSN or complete connection parameters")
	}

	// Create database driver
	var driver gorm.Dialector
	switch cfg.Type {
	case "mysql":
		driver = gormMySQL.Open(dsn)
	case "postgres":
		driver = postgres.Open(dsn)
	case "sqlite":
		driver = sqlite.Open(dsn)
	default:
		return nil, merror.Newf("unsupported database type: %s", cfg.Type)
	}

	return driver, nil
}

// createGormConfig creates GORM configuration.
func createGormConfig(cfg *Config) *gorm.Config {
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Use singular table name
		},
	}

	// Configure logger
	if cfg.Logger != nil {
		// Set log level
		logLevel := gormLogger.Warn
		switch cfg.Logger.GetLevel() {
		case mlog.ErrorLevel:
			logLevel = gormLogger.Error
		case mlog.WarnLevel:
			logLevel = gormLogger.Warn
		case mlog.InfoLevel, mlog.DebugLevel:
			logLevel = gormLogger.Info
		}

		// Set slow query threshold
		slowThreshold := cfg.SlowThreshold
		if slowThreshold == 0 {
			slowThreshold = 300 * time.Millisecond
		}

		// Create GORM logger
		gormLogger := NewGormLogger(
			cfg.Logger,
			WithLogLevel(logLevel),
			WithSlowThreshold(slowThreshold),
			WithSkipErrRecordNotFound(true),
		)
		gormConfig.Logger = gormLogger
	}

	return gormConfig
}

// configureConnectionPool sets up database connection pool.
func configureConnectionPool(db *gorm.DB, cfg *Config) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnection)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnection)
	sqlDB.SetConnMaxIdleTime(cfg.MaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

	return nil
}

// configureReplicas sets up database replicas.
func configureReplicas(db *gorm.DB, cfg *Config) error {
	if len(cfg.Replicas) == 0 {
		return nil
	}

	replicas := make([]gorm.Dialector, len(cfg.Replicas))
	for i, replicaCfg := range cfg.Replicas {
		replicaCfg = mergeReplicaConfig(cfg, replicaCfg)
		driver, err := createDriver(&replicaCfg)
		if err != nil {
			return merror.Wrapf(err, "invalid database replica configuration at index %d", i)
		}
		replicas[i] = driver
	}

	resolver := dbresolver.Register(dbresolver.Config{
		Replicas:          replicas,
		Policy:            dbresolver.RandomPolicy{}, // use random policy
		TraceResolverMode: true,
	})
	if err := db.Use(resolver); err != nil {
		if cfg.Logger != nil {
			cfg.Logger.With(mlog.String(maltose.COMPONENT, "mdb")).Errorf(context.Background(), err, "Failed to configure db resolver")
		}
		return merror.Wrap(err, "failed to configure db resolver")
	}

	return nil
}

// mergeReplicaConfig inherits connection fields omitted by a replica.
func mergeReplicaConfig(primary *Config, replica Config) Config {
	if replica.Type == "" {
		replica.Type = primary.Type
	}
	if replica.Port == "" {
		replica.Port = primary.Port
	}
	if replica.User == "" {
		replica.User = primary.User
	}
	if replica.Password == "" {
		replica.Password = primary.Password
	}
	if replica.DBName == "" {
		replica.DBName = primary.DBName
	}
	return replica
}
