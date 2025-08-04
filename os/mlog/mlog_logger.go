package mlog

import (
	"context"
	"fmt"
	"sync"
	"unsafe"

	"go.uber.org/zap"
)

// Logger is the struct for logging management.
type Logger struct {
	parent     *zap.Logger
	hooks      []Hook
	config     *Config
	level      zap.AtomicLevel
	withFields []Field
	mu         sync.RWMutex
	hookMu     sync.RWMutex
}

// New creates a new Logger instance.
func New(cfg ...*Config) *Logger {
	config := defaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	l := &Logger{
		config:     config,
		hooks:      make([]Hook, 0),
		withFields: make([]Field, 0),
	}
	// build zap logger
	l.parent, l.level = buildZapLogger(l.config)
	// add hooks
	l.AddHook(&traceHook{})
	if len(l.config.CtxKeys) > 0 {
		l.AddHook(&ctxHook{keys: l.config.CtxKeys})
	}

	return l
}

// NewWithZap creates a new Logger instance using an existing zap.Logger.
//
// This constructor is intended for advanced users or those who need customizations
// that are not supported by the standard New() constructor. When using this function:
//
// 1. You are responsible for configuring the zap.Logger (output, format, rotation, etc.)
// 2. You must manage the lifecycle of any resources associated with your zap.Logger
// 3. File rotation and cleanup are handled by your zap.Logger configuration, not by mlog
// 4. The provided config is used only for mlog-specific features (hooks, context keys, etc.)
//
// For most use cases, the standard New() constructor is recommended as it provides
// integrated file management, rotation, and cleanup functionality.
//
// Example:
//   zapLogger := zap.New(core) // Your custom zap logger
//   logger := mlog.NewWithZap(zapLogger, &mlog.Config{
//       CtxKeys: []string{"trace_id", "user_id"},
//   })
//   defer logger.Close() // This only calls zapLogger.Sync()
func NewWithZap(zapLogger *zap.Logger, cfg ...*Config) *Logger {
	config := defaultConfig()
	if len(cfg) > 0 && cfg[0] != nil {
		config = cfg[0]
	}

	l := &Logger{
		parent:     zapLogger,
		config:     config,
		level:      zap.NewAtomicLevelAt(zap.InfoLevel), // Default level, user can adjust via config
		hooks:      make([]Hook, 0),
		withFields: make([]Field, 0),
	}

	// Add hooks
	l.AddHook(&traceHook{})
	if len(l.config.CtxKeys) > 0 {
		l.AddHook(&ctxHook{keys: l.config.CtxKeys})
	}

	return l
}

// Close closes the logger and its underlying resources.
func (l *Logger) Close() error {
	// Sync flushes any buffered log entries.
	return l.parent.Sync()
}

// SetConfigWithMap sets the logger configuration using a map.
func (l *Logger) SetConfigWithMap(configMap map[string]any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Update the config struct first to parse and validate the map.
	if err := l.config.SetConfigWithMap(configMap); err != nil {
		return err
	}

	// Rebuild the zap logger with the new configuration.
	l.parent, l.level = buildZapLogger(l.config)

	// Re-apply the stored 'With' fields to the new logger instance.
	if len(l.withFields) > 0 {
		zapFields := *(*[]zap.Field)(unsafe.Pointer(&l.withFields))
		l.parent = l.parent.With(zapFields...)
	}

	l.refreshHooks()

	return nil
}

// SetConfig sets the logger configuration.
func (l *Logger) SetConfig(config *Config) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.config = config
	// Rebuild the zap logger with the new configuration.
	l.parent, l.level = buildZapLogger(l.config)

	// Re-apply the stored 'With' fields to the new logger instance.
	if len(l.withFields) > 0 {
		zapFields := *(*[]zap.Field)(unsafe.Pointer(&l.withFields))
		l.parent = l.parent.With(zapFields...)
	}

	l.refreshHooks()

	return nil
}

// With adds a field to the logger.
func (l *Logger) With(fields ...Field) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	zapFields := *(*[]zap.Field)(unsafe.Pointer(&fields))

	newZapLogger := l.parent.With(zapFields...)

	newWithFields := make([]Field, 0, len(l.withFields)+len(fields))
	newWithFields = append(newWithFields, l.withFields...)
	newWithFields = append(newWithFields, fields...)

	// 4. Return a new Logger instance that carries the full state.
	return &Logger{
		parent:     newZapLogger,
		hooks:      l.hooks,
		config:     l.config,
		level:      l.level,
		withFields: newWithFields,
	}
}

// GetConfig returns the current configuration of the logger.
func (l *Logger) GetConfig() *Config {
	return l.config
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
