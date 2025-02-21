package glog

import (
	"context"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

const DefaultName = "default"

var (
	instances = make(map[string]*Logger)
	mu        sync.RWMutex
)

type Logger struct {
	logger *logrus.Logger
}

// Instance 返回指定名称的日志实例
func Instance(name ...string) *Logger {
	instanceName := DefaultName
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}

	mu.RLock()
	if ins, ok := instances[instanceName]; ok {
		mu.RUnlock()
		return ins
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	if ins, ok := instances[instanceName]; ok {
		return ins
	}

	l := &Logger{
		logger: logrus.New(),
	}

	// 设置默认格式
	l.logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	l.logger.SetOutput(os.Stdout)

	instances[instanceName] = l
	return l
}

// SetConfigWithMap 通过 map 设置日志配置
func (l *Logger) SetConfigWithMap(config map[string]interface{}) error {
	if level, ok := config["level"]; ok {
		if lvl, err := logrus.ParseLevel(level.(string)); err == nil {
			l.logger.SetLevel(lvl)
		}
	}

	if path, ok := config["path"]; ok {
		if f, err := os.OpenFile(path.(string), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err == nil {
			l.logger.SetOutput(io.MultiWriter(os.Stdout, f))
		}
	}

	if format, ok := config["format"]; ok {
		switch format.(string) {
		case "json":
			l.logger.SetFormatter(&logrus.JSONFormatter{})
		default:
			l.logger.SetFormatter(&logrus.TextFormatter{})
		}
	}

	return nil
}

// Info 记录信息日志
func (l *Logger) Info(ctx context.Context, format string, v ...interface{}) {
	l.logger.Infof(format, v...)
}

// Error 记录错误日志
func (l *Logger) Error(ctx context.Context, format string, v ...interface{}) {
	l.logger.Errorf(format, v...)
}

// Debug 记录调试日志
func (l *Logger) Debug(ctx context.Context, format string, v ...interface{}) {
	l.logger.Debugf(format, v...)
}
