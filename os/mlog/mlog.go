package mlog

import (
	"context"
)

// ILogger is the interface for the logger.
type ILogger interface {
	Print(ctx context.Context, v ...any)                 // Print logs a message at level Info.
	Printf(ctx context.Context, format string, v ...any) // Printf logs a message at level Info.
	Debug(ctx context.Context, v ...any)                 // Debug logs a message at level Debug.
	Debugf(ctx context.Context, format string, v ...any) // Debugf logs a message at level Debug.
	Info(ctx context.Context, v ...any)                  // Info logs a message at level Info.
	Infof(ctx context.Context, format string, v ...any)  // Infof logs a message at level Info.
	Warn(ctx context.Context, v ...any)                  // Warn logs a message at level Warn.
	Warnf(ctx context.Context, format string, v ...any)  // Warnf logs a message at level Warn.
	Error(ctx context.Context, v ...any)                 // Error logs a message at level Error.
	Errorf(ctx context.Context, format string, v ...any) // Errorf logs a message at level Error.
	Fatal(ctx context.Context, v ...any)                 // Fatal logs a message at level Fatal.
	Fatalf(ctx context.Context, format string, v ...any) // Fatalf logs a message at level Fatal.
	Panic(ctx context.Context, v ...any)                 // Panic logs a message at level Panic.
	Panicf(ctx context.Context, format string, v ...any) // Panicf logs a message at level Panic.
}

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
