package mlogz

import (
	"context"
	"time"
)

// ILogger is the interface for the logger.
type ILogger interface {
	Debug(ctx context.Context, msg string)                            // Debug logs a message at level Debug.
	Debugf(ctx context.Context, format string, v ...any)              // Debugf logs a message at level Debug.
	Debugw(ctx context.Context, msg string, attrs ...Attr)            // Debugw logs a message at level Debug.
	Info(ctx context.Context, msg string)                             // Info logs a message at level Info.
	Infof(ctx context.Context, format string, v ...any)               // Infof logs a message at level Info.
	Infow(ctx context.Context, msg string, attrs ...Attr)             // Infow logs a message at level Info.
	Warn(ctx context.Context, msg string)                             // Warn logs a message at level Warn.
	Warnf(ctx context.Context, format string, v ...any)               // Warnf logs a message at level Warn.
	Warnw(ctx context.Context, msg string, attrs ...Attr)             // Warnw logs a message at level Warn.
	Error(ctx context.Context, err error, msg string)                 // Error logs a message at level Error.
	Errorf(ctx context.Context, err error, format string, v ...any)   // Errorf logs a message at level Error.
	Errorw(ctx context.Context, err error, msg string, attrs ...Attr) // Errorw logs a message at level Error.
	Fatal(ctx context.Context, err error, msg string)                 // Fatal logs a message at level Fatal.
	Fatalf(ctx context.Context, err error, format string, v ...any)   // Fatalf logs a message at level Fatal.
	Fatalw(ctx context.Context, err error, msg string, attrs ...Attr) // Fatalw logs a message at level Fatal.
	Panic(ctx context.Context, err error, msg string)                 // Panic logs a message at level Panic.
	Panicf(ctx context.Context, err error, format string, v ...any)   // Panicf logs a message at level Panic.
	Panicw(ctx context.Context, err error, msg string, attrs ...Attr) // Panicw logs a message at level Panic.
}

const (
	defaultFile       = "logs/app.log"
	defaultTimeFormat = time.DateTime
	defaultFormat     = "json"
	defaultLevel      = InfoLevel
)

// Fields is a map of string keys to any values.
type Fields map[string]any

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
