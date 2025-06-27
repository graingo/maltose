package mlogs

import (
	"context"
)

type Loggerface interface {
	Debug(context.Context, string, ...Attr)
	Info(context.Context, string, ...Attr)
	Warn(context.Context, string, ...Attr)
	Error(context.Context, error, string, ...Attr)
	Fatal(context.Context, error, string, ...Attr)
	Panic(context.Context, error, string, ...Attr)
	Debugf(context.Context, string, ...any)
	Infof(context.Context, string, ...any)
	Warnf(context.Context, string, ...any)
	Errorf(context.Context, error, string, ...any)
	Fatalf(context.Context, error, string, ...any)
	Panicf(context.Context, error, string, ...any)
}

var (
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
