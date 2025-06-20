package mlog

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is the struct for logging management.
type Logger struct {
	parent *logrus.Logger
	config Config
}

const (
	defaultPath       = "logs"
	defaultFile       = "{Y}-{m}-{d}.log"
	defaultTimeFormat = time.DateTime
	defaultFormat     = "text"
	defaultLevel      = InfoLevel
)

// New creates a new Logger instance.
func New() *Logger {
	config := DefaultConfig()
	l := &Logger{
		parent: logrus.New(),
		config: config,
	}
	l.SetConfig(config)

	// Add default hooks
	l.AddHook(&traceHook{})

	// Add context hook if there are keys configured
	if len(config.CtxKeys) > 0 {
		l.AddHook(&ctxHook{keys: config.CtxKeys})
	}

	return l
}

// Print prints `v` with newline using fmt.Sprintln.
func (l *Logger) Print(ctx context.Context, v ...any) {
	l.parent.WithContext(ctx).Print(v...)
}

// Printf prints `v` with format `format` using fmt.Sprintf.
func (l *Logger) Printf(ctx context.Context, format string, v ...any) {
	l.parent.WithContext(ctx).Printf(format, v...)
}

// Debug prints the logging content with [DEBUG] header and newline.
func (l *Logger) Debug(ctx context.Context, v ...any) {
	l.parent.WithContext(ctx).Debug(v...)
}

// Debugf prints the logging content with [DEBUG] header and format `format`.
func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.parent.WithContext(ctx).Debugf(format, v...)
}

// Info prints the logging content with [INFO] header and newline.
func (l *Logger) Info(ctx context.Context, v ...any) {
	l.parent.WithContext(ctx).Info(v...)
}

// Infof prints the logging content with [INFO] header and format `format`.
func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.parent.WithContext(ctx).Infof(format, v...)
}

// Warn prints the logging content with [WARN] header and newline.
func (l *Logger) Warn(ctx context.Context, v ...any) {
	l.parent.WithContext(ctx).Warn(v...)
}

// Warnf prints the logging content with [WARN] header and format `format`.
func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.parent.WithContext(ctx).Warnf(format, v...)
}

// Error prints the logging content with [ERROR] header and newline.
func (l *Logger) Error(ctx context.Context, v ...any) {
	l.parent.WithContext(ctx).Error(v...)
}

// Errorf prints the logging content with [ERROR] header and format `format`.
func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	l.parent.WithContext(ctx).Errorf(format, v...)
}

// Fatal prints the logging content with [FATAL] header and newline.
func (l *Logger) Fatal(ctx context.Context, v ...any) {
	l.parent.WithContext(ctx).Fatal(v...)
}

// Fatalf prints the logging content with [FATAL] header and format `format`.
func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	l.parent.WithContext(ctx).Fatalf(format, v...)
}

// Panic prints the logging content with [PANIC] header and newline.
func (l *Logger) Panic(ctx context.Context, v ...any) {
	l.parent.WithContext(ctx).Panic(v...)
}

// Panicf prints the logging content with [PANIC] header and format `format`.
func (l *Logger) Panicf(ctx context.Context, format string, v ...any) {
	l.parent.WithContext(ctx).Panicf(format, v...)
}
