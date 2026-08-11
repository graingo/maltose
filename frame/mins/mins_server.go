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

// Server returns an HTTP server instance from the default scope.
func Server(name ...string) *mhttp.Server {
	return defaultScope.Server(name...)
}

// Server returns an HTTP server instance owned by the scope.
func (s *Scope) Server(name ...string) *mhttp.Server {
	var (
		ctx          = context.Background()
		instanceName = mhttp.DefaultServerName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameServer, instanceName)

	instance := s.serverInstances.GetOrSetFunc(instanceKey, func() any {
		server := mhttp.New()

		// Apply server settings when the scope configuration is available.
		if s.Config().Available(ctx) {
			configMap, err := s.Config().Data(ctx)
			if err != nil {
				panic(merror.NewCodef(mcode.CodeMissingConfiguration, `retrieve config data map failed: %v`, err))
			}

			if serverConfigNode, ok := configMap[configNodeNameServer]; ok {
				globalConfigMap := mustConfigMap(serverConfigNode, configNodeNameServer)

				var serverConfigMap map[string]any
				// try to get instance specific config
				if instanceConfig, ok := globalConfigMap[instanceName]; ok {
					serverConfigMap = mustConfigMap(instanceConfig, fmt.Sprintf("%s.%s", configNodeNameServer, instanceName))
				} else if defaultConfig, ok := globalConfigMap["default"]; ok {
					// try to get default instance config
					serverConfigMap = mustConfigMap(defaultConfig, configNodeNameServer+".default")
				} else if len(globalConfigMap) > 0 {
					// use flat structure config
					serverConfigMap = globalConfigMap
				}

				if len(serverConfigMap) > 0 {
					if err := server.SetConfigWithMap(serverConfigMap); err != nil {
						panic(merror.NewCodef(mcode.CodeInvalidConfiguration, "set server config failed: %v", err))
					}

					// Prefer the server logger configuration, then the scope logger.
					var loggerConfigMap map[string]any
					if cfg, ok := serverConfigMap[configNodeNameLogger].(map[string]any); ok {
						loggerConfigMap = cfg
					} else if globalLoggerConfig, ok := configMap[configNodeNameLogger]; ok {
						loggerConfigMap = mustConfigMap(globalLoggerConfig, configNodeNameLogger)
					}

					// Attach a dedicated logger when configured.
					if len(loggerConfigMap) > 0 {
						serverLogger := mlog.New()
						if err := serverLogger.SetConfigWithMap(loggerConfigMap); err != nil {
							panic(merror.NewCodef(mcode.CodeInvalidConfiguration, "set server logger config failed: %v", err))
						}
						server.SetLogger(serverLogger)
					} else {
						// Otherwise, share the scope's default logger.
						server.SetLogger(s.Log())
					}
				}
			}
		}
		// Preserve the requested name when no explicit name is configured.
		if instanceName != mhttp.DefaultServerName {
			server.SetServerName(instanceName)
		}
		return server
	})

	return instance.(*mhttp.Server)
}
