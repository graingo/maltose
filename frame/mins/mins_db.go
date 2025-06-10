package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/database/mdb"
	"github.com/graingo/maltose/os/mlog"
)

const (
	configNodeNameDB = "database" // config node name for database
)

func DB(name ...string) *mdb.DB {
	var (
		ctx          = context.Background()
		instanceName = mdb.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameDB, instanceName)

	// get or create db instance
	instance := globalInstances.GetOrSetFunc(instanceKey, func() any {
		// use default config
		dbConfig := mdb.DefaultConfig()

		// if config is available, read db config from config
		if Config().Available(ctx) {
			var (
				loggerConfigMap   map[string]any
				globalConfigMap   map[string]any
				databaseConfigMap map[string]any
			)

			// get global config
			configMap, err := Config().Data(ctx)
			if err != nil {
				panic(fmt.Errorf("retrieve config data map failed: %+v", err))
			}

			// try to get db config
			if configMap[configNodeNameDB] != nil {
				globalConfigMap = configMap[configNodeNameDB].(map[string]any)

				// try to get specific instance config
				if instanceConfig, ok := globalConfigMap[instanceName]; ok {
					databaseConfigMap = instanceConfig.(map[string]any)
				} else if defaultConfig, ok := globalConfigMap["default"]; ok {
					// try to get default instance config
					databaseConfigMap = defaultConfig.(map[string]any)
				} else if len(globalConfigMap) > 0 {
					// use flat structure config
					databaseConfigMap = globalConfigMap
				}

				// apply db config
				if len(databaseConfigMap) > 0 {
					// basic config
					dbConfig.SetConfigWithMap(databaseConfigMap)

					// check current config for logger node
					if loggerConfig, ok := databaseConfigMap[configNodeNameLogger]; ok {
						// specific instance logger config
						loggerConfigMap = loggerConfig.(map[string]any)
					} else if globalLoggerConfig, ok := configMap[configNodeNameLogger]; ok {
						// global logger config
						loggerConfigMap = globalLoggerConfig.(map[string]any)
					}

					// apply logger config
					if len(loggerConfigMap) > 0 {
						dbLogger := mlog.New()
						if err := dbLogger.SetConfigWithMap(loggerConfigMap); err != nil {
							panic(fmt.Errorf("set db logger config failed: %+v", err))
						}
						dbConfig.SetLogger(dbLogger)
					}
				}
			}
		}

		// create db instance with config
		db, err := mdb.NewWithConfig(dbConfig)
		if err != nil {
			panic(err)
		}
		return db
	})

	return instance.(*mdb.DB)
}
