package mlog

import (
	"context"
	"sync"
)

const (
	DefaultName = "default"
)

// ILogger 日志接口
type ILogger interface {
	Print(ctx context.Context, v ...any)
	Printf(ctx context.Context, format string, v ...any)
	Debug(ctx context.Context, v ...any)
	Debugf(ctx context.Context, format string, v ...any)
	Info(ctx context.Context, v ...any)
	Infof(ctx context.Context, format string, v ...any)
	Warn(ctx context.Context, v ...any)
	Warnf(ctx context.Context, format string, v ...any)
	Error(ctx context.Context, v ...any)
	Errorf(ctx context.Context, format string, v ...any)
	Fatal(ctx context.Context, v ...any)
	Fatalf(ctx context.Context, format string, v ...any)
	Panic(ctx context.Context, v ...any)
	Panicf(ctx context.Context, format string, v ...any)
}

var (
	// 确保 logger 实现了 ILogger 接口
	_ ILogger = &Logger{}

	// 默认日志实例
	defaultLogger = New()

	// localInstances 使用 sync.Map 存储配置实例
	localInstances sync.Map
)

// DefaultLogger 返回默认日志实例
func DefaultLogger() *Logger {
	return defaultLogger
}

// SetDefaultLogger 设置默认日志实例
func SetDefaultLogger(logger *Logger) {
	defaultLogger = logger
}

func Instance(name ...string) *Logger {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}

	v, _ := localInstances.LoadOrStore(key, func() any {
		return New()
	}())
	return v.(*Logger)
}
