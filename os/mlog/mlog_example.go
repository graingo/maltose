package mlog

import (
	"context"

	"github.com/sirupsen/logrus"
)

// ExampleHooks 展示如何使用日志钩子功能
func ExampleHooks() {
	logger := Instance()

	// 添加多个级别的钩子
	logger.AddHooks([]logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
	}, func(entry *logrus.Entry) error {
		// 这里可以发送日志到聚合系统
		// 例如: sendToLogAggregator(entry)
		return nil
	})

	// 使用日志
	ctx := context.Background()
	logger.Error(ctx, "database connection failed")
	logger.Warn(ctx, "high memory usage: %d%%", 85)

	// 清理钩子
	logger.RemoveHooks(logrus.WarnLevel)
	// 或者清除所有钩子
	logger.ClearHooks()

	// Output:
}
