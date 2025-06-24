package mlog

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger is the struct for logging management.
type Logger struct {
	parent *logrus.Logger
	config *Config
}

const (
	defaultPath       = "logs"
	defaultFile       = "{Y}-{m}-{d}.log"
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

// extractFieldsFromArgs separates a map of fields from other logging arguments.
func (l *Logger) extractFieldsFromArgs(v []any) (logrus.Fields, []any) {
	var fields logrus.Fields
	var others []any
	for _, arg := range v {
		if m, ok := arg.(map[string]any); ok {
			if fields == nil {
				fields = logrus.Fields{}
			}
			for k, v := range m {
				fields[k] = v
			}
		} else {
			others = append(others, arg)
		}
	}
	return fields, others
}

// extractFieldsFromArgs separates a map of fields from other logging arguments.
func (l *Logger) extractFieldsFromArgs(v []any) (logrus.Fields, []any) {
	var fields logrus.Fields
	var others []any
	for _, arg := range v {
		if m, ok := arg.(map[string]any); ok {
			if fields == nil {
				fields = logrus.Fields{}
			}
			for k, v := range m {
				fields[k] = v
			}
		} else {
			others = append(others, arg)
		}
	}
	return fields, others
}

// withFields returns a new logrus.Entry with the given context and fields.
func (l *Logger) withFields(ctx context.Context, fields logrus.Fields) *logrus.Entry {
	entry := l.parent.WithContext(ctx)
	if len(fields) > 0 {
		entry = entry.WithFields(fields)
	}
	return entry
}

// Print prints `v` with newline using fmt.Sprintln.
func (l *Logger) Print(ctx context.Context, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Print(others...)
}

// Printf prints `v` with format `format` using fmt.Sprintf.
func (l *Logger) Printf(ctx context.Context, format string, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Printf(format, others...)
}

// Debug prints the logging content with [DEBUG] header and newline.
func (l *Logger) Debug(ctx context.Context, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Debug(others...)
}

// Debugf prints the logging content with [DEBUG] header and format `format`.
func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Debugf(format, others...)
}

// Info prints the logging content with [INFO] header and newline.
func (l *Logger) Info(ctx context.Context, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Info(others...)
}

// Infof prints the logging content with [INFO] header and format `format`.
func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Infof(format, others...)
}

// Warn prints the logging content with [WARN] header and newline.
func (l *Logger) Warn(ctx context.Context, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Warn(others...)
}

// Warnf prints the logging content with [WARN] header and format `format`.
func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Warnf(format, others...)
}

// Error prints the logging content with [ERROR] header and newline.
func (l *Logger) Error(ctx context.Context, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Error(others...)
}

// Errorf prints the logging content with [ERROR] header and format `format`.
func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Errorf(format, others...)
}

// Fatal prints the logging content with [FATAL] header and newline.
func (l *Logger) Fatal(ctx context.Context, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Fatal(others...)
}

// Fatalf prints the logging content with [FATAL] header and format `format`.
func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Fatalf(format, others...)
}

// Panic prints the logging content with [PANIC] header and newline.
func (l *Logger) Panic(ctx context.Context, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Panic(others...)
}

// Panicf prints the logging content with [PANIC] header and format `format`.
func (l *Logger) Panicf(ctx context.Context, format string, v ...any) {
	fields, others := l.extractFieldsFromArgs(v)
	l.withFields(ctx, fields).Panicf(format, others...)
}
