package mlog

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	logger *logrus.Logger
	field  logrus.Fields
}

// New 创建新的日志实例
func New() *Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.DateTime,
		FullTimestamp:   true,
	})

	return &Logger{
		logger: logger,
		field:  logrus.Fields{},
	}
}

// SetConfigWithMap 通过 map 设置日志配置
func (l *Logger) SetConfigWithMap(config map[string]any) error {
	if v, ok := config["level"]; ok {
		if lvl, err := logrus.ParseLevel(cast.ToString(v)); err == nil {
			l.logger.SetLevel(lvl)
		}
	}

	if v, ok := config["path"]; ok {
		if f, err := os.OpenFile(cast.ToString(v), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err == nil {
			l.logger.SetOutput(io.MultiWriter(os.Stdout, f))
		}
	}

	timeFormat := time.DateTime
	if v, ok := config["time_format"]; ok {
		timeFormat = cast.ToString(v)
	}

	if format, ok := config["format"]; ok {
		switch cast.ToString(format) {
		case "json":
			l.logger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: timeFormat,
			})
		default:
			l.logger.SetFormatter(&logrus.TextFormatter{
				TimestampFormat: timeFormat,
				FullTimestamp:   true,
			})
		}
	}

	return nil
}

func (l *Logger) Field() logrus.Fields {
	return l.field
}

func (l *Logger) SetField(field logrus.Fields) {
	l.field = field
}

func (l *Logger) getEntry(ctx context.Context) *logrus.Entry {
	fields := l.field
	if ctx != nil {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
			fields["trace_id"] = spanCtx.TraceID().String()
			fields["span_id"] = spanCtx.SpanID().String()
		}
	}
	return l.logger.WithFields(fields)
}

func (l *Logger) Print(ctx context.Context, v ...any) {
	l.getEntry(ctx).Info(v...)
}

func (l *Logger) Printf(ctx context.Context, format string, v ...any) {
	l.getEntry(ctx).Infof(format, v...)
}

func (l *Logger) Debug(ctx context.Context, v ...any) {
	l.getEntry(ctx).Debug(v...)
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.getEntry(ctx).Debugf(format, v...)
}

func (l *Logger) Info(ctx context.Context, v ...any) {
	l.getEntry(ctx).Info(v...)
}

func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.getEntry(ctx).Infof(format, v...)
}

func (l *Logger) Warn(ctx context.Context, v ...any) {
	l.getEntry(ctx).Warn(v...)
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.getEntry(ctx).Warnf(format, v...)
}

func (l *Logger) Error(ctx context.Context, v ...any) {
	l.getEntry(ctx).Error(v...)
}

func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	l.getEntry(ctx).Errorf(format, v...)
}

func (l *Logger) Fatal(ctx context.Context, v ...any) {
	l.getEntry(ctx).Fatal(v...)
}

func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	l.getEntry(ctx).Fatalf(format, v...)
}

func (l *Logger) Panic(ctx context.Context, v ...any) {
	l.getEntry(ctx).Panic(v...)
}

func (l *Logger) Panicf(ctx context.Context, format string, v ...any) {
	l.getEntry(ctx).Panicf(format, v...)
}
