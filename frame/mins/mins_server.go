package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/net/mhttp"
)

const (
	configNodeNameServer = "server" // config node name for server
)

func Server(name ...string) *mhttp.Server {
	var (
		ctx          = context.Background()
		instanceName = mhttp.DefaultServerName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameServer, instanceName)

	instance := globalInstances.GetOrSetFunc(instanceKey, func() any {
		// initialize server
		server := mhttp.New()

		// if config is available, read server config from config
		if Config().Available(ctx) {
			var (
				loggerConfigMap map[string]any
				globalConfigMap map[string]any
				serverConfigMap map[string]any
			)

			// get global config
			configMap, err := Config().Data(ctx)
			if err != nil {
				panic(fmt.Errorf("retrieve config data map failed: %+v", err))
			}

			if configMap[configNodeNameServer] != "" {
				globalConfigMap = configMap[configNodeNameServer].(map[string]any)

				// try to get instance specific config
				if instanceConfig, ok := globalConfigMap[instanceName]; ok {
					serverConfigMap = instanceConfig.(map[string]any)
				} else if defaultConfig, ok := globalConfigMap["default"]; ok {
					// try to get default instance config
					serverConfigMap = defaultConfig.(map[string]any)
				} else if len(globalConfigMap) > 0 {
					// use flat structure config
					serverConfigMap = globalConfigMap
				}

				// apply server config
				if len(serverConfigMap) > 0 {
					// basic config
					server.SetConfigWithMap(serverConfigMap)

					// check current config for logger node
					if cfg, ok := serverConfigMap[configNodeNameLogger].(map[string]any); ok {
						loggerConfigMap = cfg
					} else if globalLoggerConfig, ok := configMap[configNodeNameLogger]; ok {
						// try to get global logger config
						loggerConfigMap = globalLoggerConfig.(map[string]any)
					}

					// apply logger config
					if len(loggerConfigMap) > 0 {
						if err := server.Logger().SetConfigWithMap(loggerConfigMap); err != nil {
							panic(fmt.Errorf("set server logger config failed: %+v", err))
						}
					} else {
						// if no logger config, use global logger
						server.SetLogger(Log())
					}
				}
			}
		}
		// set server name
		if instanceName != mhttp.DefaultServerName {
			server.SetServerName(instanceName)
		}
		return server
	})

	return instance.(*mhttp.Server)
}
