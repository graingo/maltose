package mlog

import (
	"context"
)

// ILogger is the interface for the logger.
type ILogger interface {
	Debugf(ctx context.Context, format string, v ...any)                // Debugf logs a message at level Debug.
	Debugw(ctx context.Context, msg string, fields ...Field)            // Debugw logs a message at level Debug.
	Infof(ctx context.Context, format string, v ...any)                 // Infof logs a message at level Info.
	Infow(ctx context.Context, msg string, fields ...Field)             // Infow logs a message at level Info.
	Warnf(ctx context.Context, format string, v ...any)                 // Warnf logs a message at level Warn.
	Warnw(ctx context.Context, msg string, fields ...Field)             // Warnw logs a message at level Warn.
	Errorf(ctx context.Context, err error, format string, v ...any)     // Errorf logs a message at level Error.
	Errorw(ctx context.Context, err error, msg string, fields ...Field) // Errorw logs a message at level Error.
	Fatalf(ctx context.Context, err error, format string, v ...any)     // Fatalf logs a message at level Fatal.
	Fatalw(ctx context.Context, err error, msg string, fields ...Field) // Fatalw logs a message at level Fatal.
	Panicf(ctx context.Context, err error, format string, v ...any)     // Panicf logs a message at level Panic.
	Panicw(ctx context.Context, err error, msg string, fields ...Field) // Panicw logs a message at level Panic.
}

const (
	defaultFile       = ""
	defaultTimeFormat = "2006-01-02T15:04:05.000"
	defaultFormat     = "text"
	defaultLevel      = InfoLevel
)

var (
	// Ensure Logger implements ILogger interface
	_ ILogger = &Logger{}

	// defaultLogger is the default logger.
	defaultLogger = New()
)

// DefaultLogger returns the default logger.
func DefaultLogger() *Logger {
	return defaultLogger
}

// SetDefaultLogger sets the default logger for package glog.
// Note that there might be concurrent safety issue if calls this function
// in different goroutines.
func SetDefaultLogger(l *Logger) {
	defaultLogger = l
}
