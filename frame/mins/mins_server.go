package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/os/mlog"
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
		server := mhttp.New()

		// if config is available, read server config from config
		if Config().Available(ctx) {
			configMap, err := Config().Data(ctx)
			if err != nil {
				panic(merror.NewCodef(mcode.CodeMissingConfiguration, `retrieve config data map failed: %v`, err))
			}

			if serverConfigNode, ok := configMap[configNodeNameServer]; ok {
				globalConfigMap := serverConfigNode.(map[string]any)

				var serverConfigMap map[string]any
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

				if len(serverConfigMap) > 0 {
					server.SetConfigWithMap(serverConfigMap)

					// check current config for logger node
					var loggerConfigMap map[string]any
					if cfg, ok := serverConfigMap[configNodeNameLogger].(map[string]any); ok {
						loggerConfigMap = cfg
					} else if globalLoggerConfig, ok := configMap[configNodeNameLogger]; ok {
						// try to get global logger config
						loggerConfigMap = globalLoggerConfig.(map[string]any)
					}

					// apply logger config
					if len(loggerConfigMap) > 0 {
						serverLogger := mlog.New()
						if err := serverLogger.SetConfigWithMap(loggerConfigMap); err != nil {
							panic(merror.NewCodef(mcode.CodeInvalidConfiguration, "set server logger config failed: %v", err))
						}
						server.SetLogger(serverLogger)
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
