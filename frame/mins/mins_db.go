package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/database/mdb"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
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
		// If config is not available, it panics.
		if !Config().Available(ctx) {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `configuration not found for DB instance "%s"`, instanceName))
		}

		// Get global config.
		configMap, err := Config().Data(ctx)
		if err != nil {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `retrieve config data map failed: %v`, err))
		}

		// Try to get db config node.
		dbConfigNode, ok := configMap[configNodeNameDB]
		if !ok {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `configuration node "%s" not found`, configNodeNameDB))
		}

		globalConfigMap := dbConfigNode.(map[string]any)

		var databaseConfigMap map[string]any
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

		if len(databaseConfigMap) == 0 {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `no configuration found for creating database for instance "%s"`, instanceName))
		}

		dbConfig, err := mdb.ConfigFromMap(databaseConfigMap)
		if err != nil {
			panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `create database config from map failed for instance "%s": %v`, instanceName, err))
		}

		// check current config for logger node
		var loggerConfigMap map[string]any
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
				panic(merror.NewCodef(mcode.CodeInvalidConfiguration, "set db logger config failed: %v", err))
			}
			dbConfig.SetLogger(dbLogger)
		} else {
			// if no logger config, use global logger
			dbConfig.SetLogger(Log())
		}

		// create db instance with config
		db, err := mdb.New(dbConfig)
		if err != nil {
			panic(err)
		}
		return db
	})

	return instance.(*mdb.DB)
}
