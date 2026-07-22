package mdb

import (
	"context"
	"database/sql"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mlog"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	config *Config
}

func New(config ...*Config) (*DB, error) {
	cfg := defaultConfig()
	// Validate config
	if len(config) > 0 && config[0] != nil {
		cfg = cloneConfig(config[0])
	}

	ctx := context.Background()

	// Create GORM config
	gormConfig := createGormConfig(cfg)

	// Create database driver
	driver, err := createDriver(cfg)
	if err != nil {
		return nil, err
	}

	// Open database connection
	db, err := gorm.Open(driver, gormConfig)
	if err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Errorf(ctx, err, "failed to connect database")
		}
		return nil, merror.Wrap(err, "failed to open database connection")
	}

	// Configure connection pool
	if err := configureConnectionPool(db, cfg); err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Errorf(ctx, err, "failed to configure database connection pool")
		}
		closeGormDB(db)
		return nil, merror.Wrap(err, "failed to configure database connection pool")
	}

	// Configure replicas
	if err := configureReplicas(db, cfg); err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Errorf(ctx, err, "failed to configure database replicas")
		}
		closeGormDB(db)
		return nil, merror.Wrap(err, "failed to configure database replicas")
	}

	// Load plugins
	for _, plugin := range cfg.Plugins {
		if err := db.Use(plugin); err != nil {
			closeGormDB(db)
			return nil, merror.Wrap(err, "failed to load database plugin")
		}
	}

	return &DB{DB: db, config: cfg}, nil
}

func closeGormDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

// Close closes the underlying connection pool.
func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// WithContext returns a new DB with the given context.
func (db *DB) WithContext(ctx context.Context) *DB {
	return &DB{
		DB:     db.DB.WithContext(ctx),
		config: db.config,
	}
}

// Transact starts a transaction with the given context.
func (db *DB) Transact(ctx context.Context, fn func(tx *DB) error) error {
	return db.TransactWithOptions(ctx, nil, fn)
}

// TransactWithOptions starts a transaction with the given context and options.
func (db *DB) TransactWithOptions(ctx context.Context, opts *sql.TxOptions, fn func(tx *DB) error) error {
	config := db.config
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&DB{DB: tx, config: config})
	}, opts)
}

// Ping checks if the database is reachable.
func (db *DB) Ping(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (db *DB) GetLogger() *mlog.Logger {
	if db.config == nil {
		return nil
	}
	return db.config.Logger
}
