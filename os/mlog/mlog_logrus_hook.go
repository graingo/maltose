package mlog

import (
	"github.com/sirupsen/logrus"
)

// LoggerHook 定义日志钩子函数类型
type LoggerHook func(entry *logrus.Entry) error

// levelHook 实现 logrus.Hook 接口
type levelHook struct {
	level logrus.Level
	hook  LoggerHook
}

func (h *levelHook) Levels() []logrus.Level {
	return []logrus.Level{h.level}
}

func (h *levelHook) Fire(entry *logrus.Entry) error {
	return h.hook(entry)
}

// AddHook 为指定日志级别添加钩子函数
func (l *Logger) AddHook(level logrus.Level, hook LoggerHook) {
	l.logger.AddHook(&levelHook{
		level: level,
		hook:  hook,
	})
}

// AddHooks 为多个日志级别添加钩子函数
func (l *Logger) AddHooks(levels []logrus.Level, hook LoggerHook) {
	for _, level := range levels {
		l.AddHook(level, hook)
	}
}

// RemoveHooks 移除指定级别的所有钩子
func (l *Logger) RemoveHooks(level logrus.Level) {
	newLogger := logrus.New()
	newLogger.SetLevel(l.logger.GetLevel())
	newLogger.SetFormatter(l.logger.Formatter)
	newLogger.SetOutput(l.logger.Out)

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

// DingTalkHook 预设的钉钉钩子
func DingTalkHook(entry *logrus.Entry) error {
	// TODO: http请求钉钉报警
	return nil
}
