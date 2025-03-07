package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/os/mlog"
)

const (
	configNodeNameLogger = "logger" // 配置文件中的日志节点名
)

// Log 返回一个 glog.Logger 实例
// 参数 name 是实例名称
func Log(name ...string) *mlog.Logger {
	var (
		ctx          = context.Background()
		instanceName = mlog.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}

	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameLogger, instanceName)
	v := globalInstances.GetOrSetFunc(instanceKey, func() any {
		// 创建日志实例
		logger := mlog.Instance(instanceName)
		// 尝试获取日志配置
		var configMap map[string]any
		// 先尝试获取特定命名的日志配置
		certainLoggerNodeName := fmt.Sprintf(`%s.%s`, configNodeNameLogger, instanceName)
		if v, _ := Config().Get(ctx, certainLoggerNodeName); v.IsNil() {
			configMap = v.Map()
		}
		// 如果特定配置不存在，则使用全局日志配置
		if len(configMap) == 0 {
			if v, _ := Config().Get(ctx, configNodeNameLogger); !v.IsEmpty() {
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
	})

	return v.(*mlog.Logger)
}
