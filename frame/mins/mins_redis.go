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

func Redis(name ...string) *mredis.Redis {
	var (
		ctx          = context.Background()
		instanceName = mredis.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameRedis, instanceName)

	// get or create db instance
	instance := redisInstances.GetOrSetFunc(instanceKey, func() any {
		// If already configured, it returns the redis instance.
		if _, ok := mredis.GetConfig(instanceName); ok {
			return mredis.Instance(instanceName)
		}

		// if config is not available, it panics.
		if !Config().Available(ctx) {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `configuration not found for redis instance "%s"`, instanceName))
		}

		var (
			redisConfigMap map[string]any
		)
		// get global config
		configMap, err := Config().Data(ctx)
		if err != nil {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `retrieve config data map failed: %+v`, err))
		}

		// try to get redis config node.
		redisConfigNode, ok := configMap[configNodeNameRedis]
		if !ok {
			panic(merror.NewCode(mcode.CodeMissingConfiguration, `no configuration found for creating redis client`))
		}

		globalConfigMap := redisConfigNode.(map[string]any)
		// try to get specific instance config.
		if instanceConfig, ok := globalConfigMap[instanceName]; ok {
			redisConfigMap = instanceConfig.(map[string]any)
		} else if defaultConfig, ok := globalConfigMap["default"]; ok {
			// try to get default instance config
			redisConfigMap = defaultConfig.(map[string]any)
		} else if len(globalConfigMap) > 0 {
			// use flat structure config
			redisConfigMap = globalConfigMap
		}

		// parse redis config map.
		if len(redisConfigMap) == 0 {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `no configuration found for creating redis client for instance "%s"`, instanceName))
		}

		redisConfig, err := mredis.ConfigFromMap(redisConfigMap)
		if err != nil {
			panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `create redis config from map failed for instance "%s": %v`, instanceName, err))
		}

		// check current config for logger node
		var loggerConfigMap map[string]any
		if loggerConfig, ok := redisConfigMap[configNodeNameLogger]; ok {
			// specific instance logger config
			loggerConfigMap = loggerConfig.(map[string]any)
		} else if globalLoggerConfig, ok := configMap[configNodeNameLogger]; ok {
			// global logger config
			loggerConfigMap = globalLoggerConfig.(map[string]any)
		}

		// apply logger config
		if len(loggerConfigMap) > 0 {
			redisLogger := mlog.New()
			if err := redisLogger.SetConfigWithMap(loggerConfigMap); err != nil {
				panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `set redis logger config failed for instance "%s": %v`, instanceName, err))
			}
			redisConfig.SetLogger(redisLogger)
		} else {
			// if no logger config, use global logger
			redisConfig.SetLogger(Log())
		}

		// create redis instance with config.
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
