package mlogz

import (
	"context"
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/graingo/maltose/internal/intlog"
	"go.uber.org/zap"
)

// Logger is the struct for logging management.
type Logger struct {
	parent     *zap.Logger
	hooks      []Hook
	config     *Config
	level      zap.AtomicLevel
	fileWriter io.WriteCloser
	mu         sync.RWMutex
}

// New creates a new Logger instance.
func New(cfg ...*Config) *Logger {
	config := defaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	l := &Logger{
		config: config,
		hooks:  make([]Hook, 0),
	}
	// build zap logger
	l.parent, l.level, l.fileWriter = buildZapLogger(l.config)
	// add hooks
	l.AddHook(&traceHook{})
	if len(l.config.CtxKeys) > 0 {
		l.AddHook(&ctxHook{keys: l.config.CtxKeys})
	}

	return l
}

// Close closes the logger and its underlying resources.
func (l *Logger) Close() error {
	var err error
	// Sync flushes any buffered log entries.
	if syncErr := l.parent.Sync(); syncErr != nil {
		err = syncErr
	}
	// Close the writer.
	if l.fileWriter != nil {
		if closeErr := l.fileWriter.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}

// SetConfigWithMap sets the logger configuration using a map.
func (l *Logger) SetConfigWithMap(configMap map[string]any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.Close(); err != nil {
		intlog.Errorf(context.Background(), "failed to close logger: %v", err)
	}

	if err := l.config.SetConfigWithMap(configMap); err != nil {
		return err
	}
	l.parent, l.level, l.fileWriter = buildZapLogger(l.config)
	l.refreshHooks()

	return nil
}

// SetConfig sets the logger configuration.
func (l *Logger) SetConfig(config *Config) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if err := l.Close(); err != nil {
		intlog.Errorf(context.Background(), "failed to close logger: %v", err)
	}

	l.config = config
	l.parent, l.level, l.fileWriter = buildZapLogger(l.config)
	l.refreshHooks()

	return nil
}

// With adds a field to the logger.
func (l *Logger) With(fields ...Field) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	zapFields := *(*[]zap.Field)(unsafe.Pointer(&fields))

	// zap's With is immutable and returns a new logger.
	newZapLogger := l.parent.With(zapFields...)

	return &Logger{
		parent:     newZapLogger,
		hooks:      l.hooks,
		config:     l.config,
		level:      l.level,
		fileWriter: l.fileWriter,
	}
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.log(ctx, DebugLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Debugw(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, DebugLevel, msg, fields...)
}

func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.log(ctx, InfoLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Infow(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, InfoLevel, msg, fields...)
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.log(ctx, WarnLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Warnw(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, WarnLevel, msg, fields...)
}

func (l *Logger) Errorf(ctx context.Context, err error, format string, v ...any) {
	l.log(ctx, ErrorLevel, fmt.Sprintf(format, v...), Err(err))
}

func (l *Logger) Errorw(ctx context.Context, err error, msg string, fields ...Field) {
	fields = append(fields, Err(err))
	l.log(ctx, ErrorLevel, msg, fields...)
}

func (l *Logger) Fatalf(ctx context.Context, err error, format string, v ...any) {
	l.log(ctx, FatalLevel, fmt.Sprintf(format, v...), Err(err))
}

func (l *Logger) Fatalw(ctx context.Context, err error, msg string, fields ...Field) {
	fields = append(fields, Err(err))
	l.log(ctx, FatalLevel, msg, fields...)
}

func (l *Logger) Panicf(ctx context.Context, err error, format string, v ...any) {
	l.log(ctx, PanicLevel, fmt.Sprintf(format, v...), Err(err))
}

func (l *Logger) Panicw(ctx context.Context, err error, msg string, fields ...Field) {
	fields = append(fields, Err(err))
	l.log(ctx, PanicLevel, msg, fields...)
}
