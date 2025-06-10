package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/database/mredis"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/internal/intlog"
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
	instance := globalInstances.GetOrSetFunc(instanceKey, func() any {
		// If already configured, it returns the redis instance.
		if _, ok := mredis.GetConfig(instanceName); ok {
			return mredis.Instance(instanceName)
		}

		var redisConfig *mredis.Config

		// if config is available, read db config from config
		if Config().Available(ctx) {
			var (
				globalConfigMap map[string]any
				redisConfigMap  map[string]any
			)

			// get global config
			configMap, err := Config().Data(ctx)
			if err != nil {
				panic(fmt.Errorf("retrieve config data map failed: %+v", err))
			}

			// try to get db config
			if configMap[configNodeNameRedis] != nil {
				globalConfigMap = configMap[configNodeNameRedis].(map[string]any)

				// try to get specific instance config
				if instanceConfig, ok := globalConfigMap[instanceName]; ok {
					redisConfigMap = instanceConfig.(map[string]any)
				} else if defaultConfig, ok := globalConfigMap["default"]; ok {
					// try to get default instance config
					redisConfigMap = defaultConfig.(map[string]any)
				} else if len(globalConfigMap) > 0 {
					// use flat structure config
					redisConfigMap = globalConfigMap
				}

				// apply db config
				if len(redisConfigMap) > 0 {
					// basic config
					if redisConfig, err := mredis.ConfigFromMap(redisConfigMap); err == nil {
						if redisClient, err := mredis.New(redisConfig); err == nil {
							return redisClient
						} else {
							panic(err)
						}
					} else {
						intlog.Printf(ctx, `missing configuration for redis %s`, instanceName)
					}
				} else {
					panic(merror.NewCode(mcode.CodeMissingConfiguration, `no configuration found for creating redis client`))
				}
			} else {
				panic(merror.NewCode(mcode.CodeMissingConfiguration, `no configuration found for creating redis client`))
			}
		}
		// create redis instance with config.
		redisClient, err := mredis.New(redisConfig)
		if err != nil {
			panic(err)
		}
		return redisClient
	})

	return instance.(*mredis.Redis)
}
