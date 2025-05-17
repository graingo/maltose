package mdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/graingo/maltose/database/mdb/internal"
	"github.com/graingo/maltose/os/mlog"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type DB struct {
	*gorm.DB
}

func New() (*DB, error) {
	return NewWithConfig(nil)
}

func NewWithConfig(cfg *Config) (*DB, error) {
	// Validate config
	if cfg == nil {
		cfg = DefaultConfig()
	}

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
		return nil, fmt.Errorf("DSN is not set, please configure DSN or complete connection parameters")
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
		return nil, fmt.Errorf("unsupported database type: " + cfg.Type)
	}

	// Create GORM config
	gormConfig := createGormConfig(cfg)

	// Open database connection
	db, err := gorm.Open(driver, gormConfig)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Errorf(context.Background(), "failed to connect database: %v", err)
		}
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	if err := configureConnectionPool(db, cfg); err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Errorf(context.Background(), "failed to configure database connection pool: %v", err)
		}
		return nil, fmt.Errorf("failed to configure database connection pool: %w", err)
	}

	// Load tracing features
	if err := internal.LoadTracing(db); err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Errorf(context.Background(), "failed to load database tracing: %v", err)
		}
		return nil, fmt.Errorf("failed to load database tracing: %w", err)
	}

	return &DB{DB: db}, nil
}

// createGormConfig creates GORM configuration
func createGormConfig(cfg *Config) *gorm.Config {
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // Use singular table name
		},
	}

	// Configure logger
	if cfg.Logger != nil {
		// Set log level
		logLevel := logger.Warn
		switch cfg.Logger.GetLevel() {
		case mlog.ErrorLevel:
			logLevel = logger.Error
		case mlog.WarnLevel:
			logLevel = logger.Warn
		case mlog.InfoLevel, mlog.DebugLevel:
			logLevel = logger.Info
		}

		// Set slow query threshold
		slowThreshold := cfg.SlowThreshold
		if slowThreshold == 0 {
			slowThreshold = 300 * time.Millisecond
		}

		// Create GORM logger
		gormLogger := internal.NewGormLogger(
			cfg.Logger,
			internal.WithLogLevel(logLevel),
			internal.WithSlowThreshold(slowThreshold),
			internal.WithSkipErrRecordNotFound(true),
		)
		gormConfig.Logger = gormLogger
	}

	return gormConfig
}

// configureConnectionPool sets up database connection pool
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

// WithContext returns a new DB with the given context.
func (db *DB) WithContext(ctx context.Context) *DB {
	return &DB{DB: db.DB.WithContext(ctx)}
}

// Transact starts a transaction with the given context.
func (db *DB) Transact(ctx context.Context, fn func(tx *DB) error) error {
	return db.TransactWithOptions(ctx, nil, fn)
}

// TransactWithOptions starts a transaction with the given context and options.
func (db *DB) TransactWithOptions(ctx context.Context, opts *sql.TxOptions, fn func(tx *DB) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&DB{DB: tx})
	}, opts)
}
