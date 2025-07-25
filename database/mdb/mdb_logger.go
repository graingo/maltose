package mdb

import (
	"context"
	"errors"
	"time"

	"github.com/graingo/maltose/os/mlog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormLogger is a custom GORM logger that integrates with mlog.
type GormLogger struct {
	logger                *mlog.Logger
	gormLogLevel          logger.LogLevel
	slowThreshold         time.Duration
	skipErrRecordNotFound bool
}

// Option is a functional option for configuring the GormLogger.
type Option func(*GormLogger)

// WithSlowThreshold sets the slow query threshold.
func WithSlowThreshold(threshold time.Duration) Option {
	return func(l *GormLogger) {
		l.slowThreshold = threshold
	}
}

// WithSkipErrRecordNotFound sets whether to skip ErrRecordNotFound errors.
func WithSkipErrRecordNotFound(skip bool) Option {
	return func(l *GormLogger) {
		l.skipErrRecordNotFound = skip
	}
}

// WithLogLevel sets the GORM log level.
func WithLogLevel(level logger.LogLevel) Option {
	return func(l *GormLogger) {
		l.gormLogLevel = level
	}
}

// NewGormLogger creates a new GormLogger.
func NewGormLogger(mlogger *mlog.Logger, opts ...Option) *GormLogger {
	l := &GormLogger{
		logger:                mlogger,
		gormLogLevel:          logger.Warn,
		slowThreshold:         200 * time.Millisecond,
		skipErrRecordNotFound: true,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// LogMode returns a new logger with a different log level.
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.gormLogLevel = level
	return &newLogger
}

// Info logs an info message.
func (l *GormLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.gormLogLevel >= logger.Info {
		l.logger.Infof(ctx, msg, args...)
	}
}

// Warn logs a warning message.
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.gormLogLevel >= logger.Warn {
		l.logger.Warnf(ctx, msg, args...)
	}
}

// Error logs an error message.
func (l *GormLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.gormLogLevel >= logger.Error {
		l.logger.Errorf(ctx, nil, msg, args...)
	}
}

// Trace logs a SQL query.
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.gormLogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := mlog.Fields{
		mlog.String("sql", sql),
		mlog.Int64("rows", rows),
		mlog.Float64("elapsed_ms", float64(elapsed.Nanoseconds())/1e6),
	}

	switch {
	case err != nil && l.gormLogLevel >= logger.Error:
		if errors.Is(err, gorm.ErrRecordNotFound) && l.skipErrRecordNotFound {
			if l.gormLogLevel >= logger.Info {
				l.logger.Infow(ctx, "sql not found", fields...)
			}
			return
		}
		l.logger.Errorw(ctx, err, "sql error", fields...)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.gormLogLevel >= logger.Warn:
		l.logger.Warnw(ctx, "sql slow", fields...)
	case l.gormLogLevel >= logger.Info:
		l.logger.Infow(ctx, "sql trace", fields...)
	}
}
