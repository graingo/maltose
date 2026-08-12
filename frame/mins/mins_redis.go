package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/database/mredis"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mlog"
)

const (
	configNodeNameRedis = "redis" // config node name for redis
)

// Redis returns a Redis instance from the default scope.
func Redis(name ...string) *mredis.Redis {
	return defaultScope.Redis(name...)
}

// Redis returns a Redis instance owned by the scope.
func (s *Scope) Redis(name ...string) *mredis.Redis {
	var (
		ctx          = context.Background()
		instanceName = mredis.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameRedis, instanceName)

	// Create each named instance at most once within the scope.
	instance := s.redisInstances.GetOrSetFunc(instanceKey, func() any {
		// Preserve package-global Redis instances for the default scope only.
		if _, ok := mredis.GetConfig(instanceName); s.useGlobalComponents && ok {
			return mredis.Instance(instanceName)
		}

		// Redis initialization requires an available configuration source.
		if !s.Config().Available(ctx) {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `configuration not found for redis instance "%s"`, instanceName))
		}

		var (
			redisConfigMap map[string]any
		)
		// Read the complete scope configuration once for component fallbacks.
		configMap, err := s.Config().Data(ctx)
		if err != nil {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `retrieve config data map failed: %+v`, err))
		}

		// Locate the Redis configuration node.
		redisConfigNode, ok := configMap[configNodeNameRedis]
		if !ok {
			panic(merror.NewCode(mcode.CodeMissingConfiguration, `no configuration found for creating redis client`))
		}

		globalConfigMap := mustConfigMap(redisConfigNode, configNodeNameRedis)
		// try to get specific instance config.
		if instanceConfig, ok := globalConfigMap[instanceName]; ok {
			redisConfigMap = mustConfigMap(instanceConfig, fmt.Sprintf("%s.%s", configNodeNameRedis, instanceName))
		} else if defaultConfig, ok := globalConfigMap["default"]; ok {
			// try to get default instance config
			redisConfigMap = mustConfigMap(defaultConfig, configNodeNameRedis+".default")
		} else if len(globalConfigMap) > 0 {
			// use flat structure config
			redisConfigMap = globalConfigMap
		}

		// Convert the selected node into a Redis configuration.
		if len(redisConfigMap) == 0 {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `no configuration found for creating redis client for instance "%s"`, instanceName))
		}

		redisConfig, err := mredis.ConfigFromMap(redisConfigMap)
		if err != nil {
			panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `create redis config from map failed for instance "%s": %v`, instanceName, err))
		}

		// Prefer the instance logger configuration, then the scope logger.
		var loggerConfigMap map[string]any
		if loggerConfig, ok := redisConfigMap[configNodeNameLogger]; ok {
			loggerConfigMap = mustConfigMap(loggerConfig, fmt.Sprintf("%s.%s.%s", configNodeNameRedis, instanceName, configNodeNameLogger))
		} else if globalLoggerConfig, ok := configMap[configNodeNameLogger]; ok {
			loggerConfigMap = mustConfigMap(globalLoggerConfig, configNodeNameLogger)
		}

		// Attach a dedicated logger when configured.
		if len(loggerConfigMap) > 0 {
			redisLogger := mlog.New()
			if err := redisLogger.SetConfigWithMap(loggerConfigMap); err != nil {
				panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `set redis logger config failed for instance "%s": %v`, instanceName, err))
			}
			redisConfig.SetLogger(redisLogger)
		} else {
			// Otherwise, share the scope's default logger.
			redisConfig.SetLogger(s.Log())
		}

		// Create the client only after configuration is complete.
		redisClient, err := mredis.New(redisConfig)
		if err != nil {
			panic(err)
		}
		return redisClient
	})

	if instance == nil {
		return nil
	}
	return instance.(*mredis.Redis)
}

// TryRedis returns a Redis instance or an initialization error.
// Unlike Redis, it does not panic when configuration or client setup fails.
func TryRedis(name ...string) (client *mredis.Redis, err error) {
	return defaultScope.TryRedis(name...)
}

// TryRedis returns a scoped Redis instance or an initialization error.
func (s *Scope) TryRedis(name ...string) (client *mredis.Redis, err error) {
	defer recoverAsError(&err)
	return s.Redis(name...), nil
}
