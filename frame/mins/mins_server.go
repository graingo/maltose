package mins

import (
	"context"
	"fmt"

	"github.com/savorelle/maltose/net/mhttp"
	"github.com/spf13/cast"
)

const (
	// 配置节点名称
	configNodeNameServer = "server"
)

func Server(name ...interface{}) *mhttp.Server {
	var (
		ctx          = context.Background()
		instanceName = mhttp.DefaultServerName
		instanceKey  = fmt.Sprintf("%s.%v", frameCoreNameServer, name)
	)

	if len(name) > 0 && name[0] != "" {
		instanceName = cast.ToString(name[0])
	}

	return globalInstances.GetOrSetFunc(instanceKey, func() interface{} {
		server := mhttp.New()

		if Config().Available(ctx) {
			var (
				configMap             map[string]any
				serverConfigMap       map[string]any
				loggerServerConfigMap map[string]any
			)

			// 获取全局配置
			configMap, err := Config().Data(ctx)
			if err != nil {
				panic(fmt.Errorf("retrieve config data map failed: %+v", err))
			}

			if configMap[configNodeNameServer] != "" {
				serverConfigMap = make(map[string]any)

				// 1. 尝试获取实例特定配置
				if instanceConfig, ok := configMap[configNodeNameServer].(map[string]any)[instanceName]; ok {
					serverConfigMap = instanceConfig.(map[string]any)
				} else if defaultConfig, ok := configMap[configNodeNameServer].(map[string]any)["default"]; ok {
					// 2. 尝试获取默认实例配置
					serverConfigMap = defaultConfig.(map[string]any)
				} else {
					// 3. 使用扁平结构配置
					serverConfigMap = configMap[configNodeNameServer].(map[string]any)
				}

				// 应用服务器配置
				if len(serverConfigMap) > 0 {
					// 基础配置
					server.SetConfig(serverConfigMap)

					// 日志配置处理
					loggerServerConfigMap = make(map[string]any)

					// 1. 先检查当前配置中的logger节点
					if cfg, ok := serverConfigMap[configNodeNameLogger].(map[string]any); ok {
						loggerServerConfigMap = cfg
					} else {
						// 2. 尝试获取全局logger配置
						if cfg, ok := configMap["logger"].(map[string]any); ok {
							loggerServerConfigMap = cfg
						}
					}

					if len(loggerServerConfigMap) > 0 {
						if err := server.Logger().SetConfigWithMap(loggerServerConfigMap); err != nil {
							panic(err)
						}
					}
				}
			}
		}

		// 设置服务器名称
		if instanceName != mhttp.DefaultServerName {
			server.SetServerName(instanceName)
		}
		return server
	}).(*mhttp.Server)
}
