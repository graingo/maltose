package mlogs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	logger *slog.Logger
	config *Config
}

func New(cfg ...*Config) *Logger {
	config := defaultConfig()
	if len(cfg) > 0 {
		config = cfg[0]
	}

	l := &Logger{
		config: config,
	}
	l.logger = l.buildSlogLogger()

	return l
}

func (l *Logger) buildSlogLogger() *slog.Logger {
	var writers []io.Writer
	if l.config.Stdout {
		writers = append(writers, os.Stdout)
	}
	if l.config.Path != "" && l.config.File != "" {
		writers = append(writers, &lumberjack.Logger{
			Filename:   filepath.Join(l.config.Path, l.config.File),
			MaxSize:    500,
			MaxAge:     l.config.AutoClean,
			MaxBackups: 3,
			LocalTime:  true,
			Compress:   false,
		})
	}
	writer := io.MultiWriter(writers...)

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     l.config.Level.toSlogLevel(),
	}

	var baseHandler slog.Handler
	if l.config.Format == "json" {
		baseHandler = slog.NewJSONHandler(writer, opts)
	} else {
		baseHandler = slog.NewTextHandler(writer, opts)
	}

	var handler slog.Handler = NewTraceHandler(baseHandler)
	if len(l.config.CtxKeys) > 0 {
		handler = NewCtxHandler(handler, l.config.CtxKeys)
	}
	for _, hook := range l.config.Hooks {
		handler = hook.New(handler)
	}

	return slog.New(handler)
}

func (l *Logger) SetConfig(config *Config) error {
	l.config = config
	l.logger = l.buildSlogLogger()
	return nil
}

// --- Contextual Loggers ---

func (l *Logger) clone() *Logger {
	newLogger := *l
	return &newLogger
}

func (l *Logger) With(attrs ...Attr) *Logger {
	newLogger := l.clone()
	// slog.Logger.With expects []any, so we must convert.
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	newLogger.logger = l.logger.With(args...)
	return newLogger
}

// --- Direct Logging Methods ---

func (l *Logger) log(level Level, ctx context.Context, msg string, attrs ...Attr) {
	if l.logger.Enabled(ctx, level.toSlogLevel()) {
		l.logger.LogAttrs(ctx, level.toSlogLevel(), msg, attrs...)
	}

	if level >= PanicLevel {
		panic(fmt.Sprintf("%s: %v", msg, attrs))
	}
	if level >= FatalLevel {
		os.Exit(1)
	}
}

func (l *Logger) Debug(ctx context.Context, msg string, attrs ...Attr) {
	l.log(DebugLevel, ctx, msg, attrs...)
}

func (l *Logger) Info(ctx context.Context, msg string, attrs ...Attr) {
	l.log(InfoLevel, ctx, msg, attrs...)
}

func (l *Logger) Warn(ctx context.Context, msg string, attrs ...Attr) {
	l.log(WarnLevel, ctx, msg, attrs...)
}

func (l *Logger) Error(ctx context.Context, err error, msg string, attrs ...Attr) {
	attrs = append(attrs, Err(err))
	l.log(ErrorLevel, ctx, msg, attrs...)
}

func (l *Logger) Fatal(ctx context.Context, err error, msg string, attrs ...Attr) {
	attrs = append(attrs, Err(err))
	l.log(FatalLevel, ctx, msg, attrs...)
}

func (l *Logger) Panic(ctx context.Context, err error, msg string, attrs ...Attr) {
	attrs = append(attrs, Err(err))
	l.log(PanicLevel, ctx, msg, attrs...)
}

func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.log(DebugLevel, ctx, fmt.Sprintf(format, args...))
}

func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	l.log(InfoLevel, ctx, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.log(WarnLevel, ctx, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(ctx context.Context, err error, format string, args ...any) {
	l.log(ErrorLevel, ctx, fmt.Sprintf(format, args...), Err(err))
}

func (l *Logger) Fatalf(ctx context.Context, err error, format string, args ...any) {
	l.log(FatalLevel, ctx, fmt.Sprintf(format, args...), Err(err))
}

func (l *Logger) Panicf(ctx context.Context, err error, format string, args ...any) {
	l.log(PanicLevel, ctx, fmt.Sprintf(format, args...), Err(err))
}

func (l *Logger) SetHooks(hooks ...Hook) {
	l.config.Hooks = hooks
}

func (l *Logger) GetHooks() []Hook {
	return l.config.Hooks
}

func (l *Logger) AddHook(hook Hook) {
	l.config.Hooks = append(l.config.Hooks, hook)
}
