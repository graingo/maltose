package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/spf13/cast"
)

const (
	configNodeNameServer = "server" // config node name for server
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

			// get global config
			configMap, err := Config().Data(ctx)
			if err != nil {
				panic(fmt.Errorf("retrieve config data map failed: %+v", err))
			}

			if configMap[configNodeNameServer] != "" {
				serverConfigMap = make(map[string]any)

				// try to get instance specific config
				if instanceConfig, ok := configMap[configNodeNameServer].(map[string]any)[instanceName]; ok {
					serverConfigMap = instanceConfig.(map[string]any)
				} else if defaultConfig, ok := configMap[configNodeNameServer].(map[string]any)["default"]; ok {
					// try to get default instance config
					serverConfigMap = defaultConfig.(map[string]any)
				} else {
					// use flat structure config
					serverConfigMap = configMap[configNodeNameServer].(map[string]any)
				}

				// apply server config
				if len(serverConfigMap) > 0 {
					// basic config
					server.SetConfigWithMap(serverConfigMap)
					// logger config processing
					loggerServerConfigMap = make(map[string]any)

					// check current config for logger node
					if cfg, ok := serverConfigMap[configNodeNameLogger].(map[string]any); ok {
						loggerServerConfigMap = cfg
					} else {
						// try to get global logger config
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
		// set server name
		if instanceName != mhttp.DefaultServerName {
			server.SetServerName(instanceName)
		}
		return server
	}).(*mhttp.Server)
}
