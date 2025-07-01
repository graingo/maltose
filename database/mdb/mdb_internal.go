package mdb

import (
	"context"
	"fmt"
	"time"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mlog"
	"gorm.io/driver/mysql"
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
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		case "postgres":
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
				cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
		}
	}

	if dsn == "" {
		return nil, merror.New("DSN is not set, please configure DSN or complete connection parameters")
	}

	// Create database driver
	var driver gorm.Dialector
	switch cfg.Type {
	case "mysql":
		driver = mysql.Open(dsn)
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

	// Set maximum number of idle connections
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnection)

	// Set maximum number of open connections
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnection)

	// Set maximum idle time for connections
	maxIdleTime := cfg.MaxIdleTime
	if maxIdleTime <= 0 {
		maxIdleTime = time.Hour // Default 1 hour
	}
	sqlDB.SetConnMaxIdleTime(maxIdleTime)

	return nil
}

// configureReplicas sets up database replicas.
func configureReplicas(db *gorm.DB, cfg *Config) error {
	if len(cfg.Replicas) == 0 {
		return nil
	}

	replicas := make([]gorm.Dialector, len(cfg.Replicas))
	for i, replicaCfg := range cfg.Replicas {
		// Note: we need to pass the address of the replica config
		driver, err := createDriver(&replicaCfg)
		if err != nil {
			return err
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
