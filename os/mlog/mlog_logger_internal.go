package mlog

import (
	"context"
	"io"
	"os"
	"slices"
	"unsafe"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (l *Logger) refreshHooks() {
	l.RemoveHook(ctxHookName)
	if len(l.config.CtxKeys) > 0 {
		l.AddHook(&ctxHook{keys: l.config.CtxKeys})
	}
}

func buildZapLogger(config *Config) (*zap.Logger, zap.AtomicLevel, io.WriteCloser) {
	encoderCfg := zap.NewProductionEncoderConfig()

	// TimeFormat
	if config.TimeFormat != "" {
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(config.TimeFormat)
	} else {
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Encoder
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// Writer
	var fileWriter io.WriteCloser
	var err error
	writers := make([]zapcore.WriteSyncer, 0, 2)
	if config.Stdout {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}
	if config.Filepath != "" {
		fileWriter, err = newFileWriter(config.Filepath, &rotationConfig{
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

	return zapLogger, level, fileWriter
}

// log logs the message with the given level and attributes.
func (l *Logger) log(ctx context.Context, level Level, msg string, fields ...Field) {
	// Get entry from the pool.
	entry := entryPool.Get().(*Entry)
	entry.ctx = ctx
	entry.msg = msg
	entry.fields = append(entry.fields[:0], fields...)

	// Reset entry and put it back to the pool.
	defer func() {
		entry.reset()
		entryPool.Put(entry)
	}()

	// Fire hooks.
	for _, hook := range l.hooks {
		if slices.Contains(hook.Levels(), level) {
			hook.Fire(entry)
		}
	}

	// NOTE: This is a zero-cost cast that relies on mlog.Field having the same memory layout as zapcore.Field.
	// Be cautious when upgrading zap or changing mlog.Field definition.
	zapFields := *(*[]zap.Field)(unsafe.Pointer(&entry.fields))
	switch level {
	case DebugLevel:
		l.parent.Debug(entry.msg, zapFields...)
	case InfoLevel:
		l.parent.Info(entry.msg, zapFields...)
	case WarnLevel:
		l.parent.Warn(entry.msg, zapFields...)
	case ErrorLevel:
		l.parent.Error(entry.msg, zapFields...)
	case FatalLevel:
		l.parent.Fatal(entry.msg, zapFields...)
	case PanicLevel:
		l.parent.Panic(entry.msg, zapFields...)
	}
}
