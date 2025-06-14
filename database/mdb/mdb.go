package mdb

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
	config *Config
}

func New(config ...*Config) (*DB, error) {
	cfg := DefaultConfig()
	// Validate config
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	}

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

	// Configure replicas
	if err := configureReplicas(db, cfg); err != nil {
		if cfg.Logger != nil {
			cfg.Logger.Errorf(context.Background(), "failed to configure database replicas: %v", err)
		}
		return nil, fmt.Errorf("failed to configure database replicas: %w", err)
	}

	// Load plugins
	for _, plugin := range cfg.Plugins {
		if err := db.Use(plugin); err != nil {
			return nil, fmt.Errorf("failed to load database plugin: %w", err)
		}
	}

	return &DB{DB: db, config: cfg}, nil
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

// Ping checks if the database is reachable.
func (db *DB) Ping(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
