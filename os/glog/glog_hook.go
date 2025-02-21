package glog

import (
	"github.com/sirupsen/logrus"
)

// LogHook 定义日志钩子函数类型
type LogHook func(entry *logrus.Entry) error

// AddHook 为指定日志级别添加钩子函数
func (l *Logger) AddHook(level logrus.Level, hook LogHook) {
	l.logger.AddHook(&levelHook{
		level: level,
		hook:  hook,
	})
}

// AddHooks 为多个日志级别添加钩子函数
func (l *Logger) AddHooks(levels []logrus.Level, hook LogHook) {
	for _, level := range levels {
		l.AddHook(level, hook)
	}
}

// RemoveHooks 移除指定级别的所有钩子
func (l *Logger) RemoveHooks(level logrus.Level) {
	// logrus 不直接支持移除钩子，需要重新创建 logger
	newLogger := logrus.New()
	newLogger.SetLevel(l.logger.GetLevel())
	newLogger.SetFormatter(l.logger.Formatter)
	newLogger.SetOutput(l.logger.Out)

	// 复制其他级别的钩子
	for _, hook := range l.logger.Hooks {
		for _, h := range hook {
			if h.(*levelHook).level != level {
				newLogger.AddHook(h)
			}
		}
	}

	l.logger = newLogger
}

// ClearHooks 清除所有钩子
func (l *Logger) ClearHooks() {
	l.logger.Hooks = make(logrus.LevelHooks)
}

// levelHook 实现 logrus.Hook 接口
type levelHook struct {
	level logrus.Level
	hook  LogHook
}

func (h *levelHook) Levels() []logrus.Level {
	return []logrus.Level{h.level}
}

func (h *levelHook) Fire(entry *logrus.Entry) error {
	return h.hook(entry)
}

// func main() {
// 	logger := glog.Instance()

// 	// 添加错误日志文件钩子
// 	logger.AddHook(logrus.ErrorLevel, glog.NewFileHook(logrus.ErrorLevel, "error.log"))

// 	// 添加警告回调钩子
// 	logger.AddHook(logrus.WarnLevel, glog.NewCallbackHook(logrus.WarnLevel, func(msg string, fields logrus.Fields) {
// 		// 发送警告消息到监控系统
// 		metrics.SendWarning(msg, fields)
// 	}))

// 	// 添加多个级别的钩子
// 	logger.AddHooks([]logrus.Level{
// 		logrus.ErrorLevel,
// 		logrus.WarnLevel,
// 	}, func(entry *logrus.Entry) error {
// 		// 发送到日志聚合系统
// 		return sendToLogAggregator(entry)
// 	})

// 	// 使用日志
// 	ctx := context.Background()
// 	logger.Error(ctx, "database connection failed")
// 	logger.Warn(ctx, "high memory usage: %d%%", 85)

// 	// 清理钩子
// 	logger.RemoveHooks(logrus.WarnLevel)
// 	// 或者清除所有钩子
// 	logger.ClearHooks()
// }
