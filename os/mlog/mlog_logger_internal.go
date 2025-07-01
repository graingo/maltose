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
	// Fire hooks before logging.
	for _, hook := range l.hooks {
		// Fire the hook if the message level is at or above the hook's level.
		if slices.Contains(hook.Levels(), level) {
			msg, fields = hook.Fire(ctx, msg, fields)
		}
	}

	zapFields := *(*[]zap.Field)(unsafe.Pointer(&fields))
	switch level {
	case DebugLevel:
		l.parent.Debug(msg, zapFields...)
	case InfoLevel:
		l.parent.Info(msg, zapFields...)
	case WarnLevel:
		l.parent.Warn(msg, zapFields...)
	case ErrorLevel:
		l.parent.Error(msg, zapFields...)
	case FatalLevel:
		l.parent.Fatal(msg, zapFields...)
	case PanicLevel:
		l.parent.Panic(msg, zapFields...)
	}
}
