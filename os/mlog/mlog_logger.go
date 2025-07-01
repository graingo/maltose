package mlog

import (
	"context"
	"time"

	"github.com/graingo/maltose"
	"github.com/sirupsen/logrus"
)

// Logger is the struct for logging management.
type Logger struct {
	parent *logrus.Logger
	entry  *logrus.Entry
	config *Config
}

const (
	defaultFile       = "logs/app.log"
	defaultTimeFormat = time.DateTime
	defaultFormat     = "json"
	defaultLevel      = InfoLevel
)

// New creates a new Logger instance.
func New(cfg ...*Config) *Logger {
	config := defaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	l := &Logger{
		parent: logrus.New(),
		config: config,
	}
	l.loadConfig(config)

	// Add default hooks
	l.AddHook(&traceHook{})

	return l
}

// clone creates a new logger object with a copy of the current logger's properties.
func (l *Logger) clone() *Logger {
	newLogger := *l
	return &newLogger
}

// getEntry returns the current logrus entry, creating one if it doesn't exist.
func (l *Logger) getEntry(ctx context.Context) *logrus.Entry {
	entry := l.entry
	if entry == nil {
		entry = l.parent.WithContext(ctx)
	} else {
		entry = entry.WithContext(ctx)
	}
	return entry
}

// extractFieldsFromArgs separates a map of fields from other logging arguments
// and returns the remaining arguments and a new logrus.Entry with the fields.
func (l *Logger) extractFieldsFromArgs(ctx context.Context, v []any) ([]any, *logrus.Entry) {
	var others []any
	entry := l.getEntry(ctx)
	for _, arg := range v {
		if m, ok := arg.(Fields); ok {
			entry = entry.WithFields(logrus.Fields(m))
		} else if m, ok := arg.(map[string]any); ok {
			entry = entry.WithFields(m)
		} else {
			others = append(others, arg)
		}
	}
	return others, entry
}

// WithComponent adds a component field to the logger.
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithField(maltose.COMPONENT, component)
}

// WithField adds a field to the logger.
func (l *Logger) WithField(key string, value any) *Logger {
	newLogger := l.clone()
	newLogger.entry = newLogger.getEntry(context.Background()).WithField(key, value)
	return newLogger
}

// WithFields adds multiple fields to the logger.
func (l *Logger) WithFields(fields Fields) *Logger {
	newLogger := l.clone()
	newLogger.entry = newLogger.getEntry(context.Background()).WithFields(logrus.Fields(fields))
	return newLogger
}

// Print prints `v` with newline using fmt.Sprintln.
func (l *Logger) Print(ctx context.Context, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Print(others...)
}

// Printf prints `v` with format `format` using fmt.Sprintf.
func (l *Logger) Printf(ctx context.Context, format string, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Printf(format, others...)
}

// Debug prints the logging content with [DEBUG] header and newline.
func (l *Logger) Debug(ctx context.Context, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Debug(others...)
}

// Debugf prints the logging content with [DEBUG] header and format `format`.
func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Debugf(format, others...)
}

// Info prints the logging content with [INFO] header and newline.
func (l *Logger) Info(ctx context.Context, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Info(others...)
}

// Infof prints the logging content with [INFO] header and format `format`.
func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Infof(format, others...)
}

// Warn prints the logging content with [WARN] header and newline.
func (l *Logger) Warn(ctx context.Context, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Warn(others...)
}

// Warnf prints the logging content with [WARN] header and format `format`.
func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Warnf(format, others...)
}

// Error prints the logging content with [ERROR] header and newline.
func (l *Logger) Error(ctx context.Context, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Error(others...)
}

// Errorf prints the logging content with [ERROR] header and format `format`.
func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Errorf(format, others...)
}

// Fatal prints the logging content with [FATAL] header and newline.
func (l *Logger) Fatal(ctx context.Context, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Fatal(others...)
}

// Fatalf prints the logging content with [FATAL] header and format `format`.
func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Fatalf(format, others...)
}

// Panic prints the logging content with [PANIC] header and newline.
func (l *Logger) Panic(ctx context.Context, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Panic(others...)
}

// Panicf prints the logging content with [PANIC] header and format `format`.
func (l *Logger) Panicf(ctx context.Context, format string, v ...any) {
	others, entry := l.extractFieldsFromArgs(ctx, v)
	entry.Panicf(format, others...)
}
