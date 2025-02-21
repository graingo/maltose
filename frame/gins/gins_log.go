package gins

import (
	"context"
	"fmt"

	"github.com/mingzaily/maltose/os/glog"
)

const (
	frameCoreComponentNameLogger = "logger"
	configNodeNameLogger         = "logger" // 配置文件中的日志节点名
)

// Log 返回一个 glog.Logger 实例
// 参数 name 是实例名称
func Log(name ...string) *glog.Logger {
	var (
		ctx          = context.Background()
		instanceName = glog.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}

	logger := glog.Instance(instanceName)

	// 尝试获取日志配置
	var configMap map[string]interface{}

	// 先尝试获取特定命名的日志配置
	certainLoggerNodeName := fmt.Sprintf(`%s.%s`, configNodeNameLogger, instanceName)
	if v := Config().Get(ctx, certainLoggerNodeName); v.IsNil() {
		configMap = v.Map()
	}

	// 如果特定配置不存在，则使用全局日志配置
	if len(configMap) == 0 {
		if v := Config().Get(ctx, configNodeNameLogger); !v.IsEmpty() {
			configMap = v.Map()
		}
	}

	// 如果存在配置，则设置到日志实例
	if len(configMap) > 0 {
		if err := logger.SetConfigWithMap(configMap); err != nil {
			panic(err)
		}
	}

	return logger
}
