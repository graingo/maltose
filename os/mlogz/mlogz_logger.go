package mlogz

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/graingo/maltose/net/mtrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the struct for logging management.
type Logger struct {
	parent *zap.Logger
	hooks  []Hook
	config *Config
	level  Level
	mu     sync.RWMutex
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
	l.parent = buildZapLogger(l.config)
	return l
}

func buildZapLogger(config *Config) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()

	// TimeFormat
	if config.TimeFormat != "" {
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(config.TimeFormat)
	}

	// Encoder
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// Writer
	writers := make([]zapcore.WriteSyncer, 0, 2)
	if config.Stdout {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	if config.Filepath != "" {
		fileWriter, err := newFileWriter(config.Filepath, &rotationConfig{
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
		})
		if err != nil {
			panic(err)
		}
		writers = append(writers, zapcore.AddSync(fileWriter))
	}
	if len(writers) == 0 {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	writeSyncer := zapcore.NewMultiWriteSyncer(writers...)

	// Level
	level := zap.NewAtomicLevelAt(zapcore.Level(config.Level))
	// Core
	core := zapcore.NewCore(encoder, writeSyncer, level)
	// Logger
	zapLogger := zap.New(core)
	if config.Caller {
		zapLogger = zapLogger.WithOptions(zap.AddCaller())
	}

	return zapLogger
}

func (l *Logger) log(ctx context.Context, level Level, msg string, attrs ...Attr) {
	// Fire hooks before logging.
	for _, hook := range l.hooks {
		// Fire the hook if the message level is at or above the hook's level.
		if level >= hook.Level() {
			msg, attrs = hook.Fire(ctx, msg, attrs)
		}
	}

	// internal function
	// If the context is not nil, add trace and span id
	if mtrace.GetTraceID(ctx) != "" {
		attrs = append(attrs, Attr{
			Key:   "trace_id",
			Value: mtrace.GetTraceID(ctx),
		})
	}
	if mtrace.GetSpanID(ctx) != "" {
		attrs = append(attrs, Attr{
			Key:   "span_id",
			Value: mtrace.GetSpanID(ctx),
		})
	}

	fields := l.toZapFields(attrs)
	switch level {
	case DebugLevel:
		l.parent.Debug(msg, fields...)
	case InfoLevel:
		l.parent.Info(msg, fields...)
	case WarnLevel:
		l.parent.Warn(msg, fields...)
	case ErrorLevel:
		l.parent.Error(msg, fields...)
	case FatalLevel:
		l.parent.Fatal(msg, fields...)
	case PanicLevel:
		l.parent.Panic(msg, fields...)
	}
}

func (l *Logger) toZapFields(fields []Attr) []zapcore.Field {
	zapFields := make([]zapcore.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}

// AddHook adds a hook to the logger.
func (l *Logger) AddHook(hook Hook) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hooks = append(l.hooks, hook)
}

// SetConfigWithMap sets the logger configuration using a map.
func (l *Logger) SetConfigWithMap(configMap map[string]any) error {
	if err := l.config.SetConfigWithMap(configMap); err != nil {
		return err
	}
	l.parent = buildZapLogger(l.config)
	return nil
}

// SetConfig sets the logger configuration.
func (l *Logger) SetConfig(config *Config) error {
	l.config = config
	l.parent = buildZapLogger(l.config)
	return nil
}

// With adds a field to the logger.
func (l *Logger) With(attrs ...Attr) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.parent = l.parent.With(l.toZapFields(attrs)...)
	return l
}

func (l *Logger) Debug(ctx context.Context, msg string) {
	l.log(ctx, DebugLevel, msg)
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.log(ctx, DebugLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Debugw(ctx context.Context, msg string, attrs ...Attr) {
	l.log(ctx, DebugLevel, msg, attrs...)
}

func (l *Logger) Info(ctx context.Context, msg string) {
	l.log(ctx, InfoLevel, msg)
}

func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.log(ctx, InfoLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Infow(ctx context.Context, msg string, attrs ...Attr) {
	l.log(ctx, InfoLevel, msg, attrs...)
}

func (l *Logger) Warn(ctx context.Context, msg string) {
	l.log(ctx, WarnLevel, msg)
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.log(ctx, WarnLevel, fmt.Sprintf(format, v...))
}

func (l *Logger) Warnw(ctx context.Context, msg string, attrs ...Attr) {
	l.log(ctx, WarnLevel, msg, attrs...)
}

func (l *Logger) Error(ctx context.Context, err error, msg string) {
	l.log(ctx, ErrorLevel, msg, Err(err))
}

func (l *Logger) Errorf(ctx context.Context, err error, format string, v ...any) {
	l.log(ctx, ErrorLevel, fmt.Sprintf(format, v...), Err(err))
}

func (l *Logger) Errorw(ctx context.Context, err error, msg string, attrs ...Attr) {
	attrs = append(attrs, Err(err))
	l.log(ctx, ErrorLevel, msg, attrs...)
}

func (l *Logger) Fatal(ctx context.Context, err error, msg string) {
	l.log(ctx, FatalLevel, msg, Err(err))
}

func (l *Logger) Fatalf(ctx context.Context, err error, format string, v ...any) {
	l.log(ctx, FatalLevel, fmt.Sprintf(format, v...), Err(err))
}

func (l *Logger) Fatalw(ctx context.Context, err error, msg string, attrs ...Attr) {
	attrs = append(attrs, Err(err))
	l.log(ctx, FatalLevel, msg, attrs...)
}

func (l *Logger) Panic(ctx context.Context, err error, msg string) {
	l.log(ctx, PanicLevel, msg, Err(err))
}

func (l *Logger) Panicf(ctx context.Context, err error, format string, v ...any) {
	l.log(ctx, PanicLevel, fmt.Sprintf(format, v...), Err(err))
}

func (l *Logger) Panicw(ctx context.Context, err error, msg string, attrs ...Attr) {
	attrs = append(attrs, Err(err))
	l.log(ctx, PanicLevel, msg, attrs...)
}
