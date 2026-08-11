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

// DB returns a database instance from the default scope.
func DB(name ...string) *mdb.DB {
	return defaultScope.DB(name...)
}

// DB returns a database instance owned by the scope.
func (s *Scope) DB(name ...string) *mdb.DB {
	var (
		ctx          = context.Background()
		instanceName = mdb.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameDB, instanceName)

	// Create each named instance at most once within the scope.
	instance := s.dbInstances.GetOrSetFunc(instanceKey, func() any {
		// Database initialization requires an available configuration source.
		if !s.Config().Available(ctx) {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `configuration not found for DB instance "%s"`, instanceName))
		}

		// Read the complete scope configuration once for component fallbacks.
		configMap, err := s.Config().Data(ctx)
		if err != nil {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `retrieve config data map failed: %v`, err))
		}

		// Locate the database configuration node.
		dbConfigNode, ok := configMap[configNodeNameDB]
		if !ok {
			panic(merror.NewCodef(mcode.CodeMissingConfiguration, `configuration node "%s" not found`, configNodeNameDB))
		}

		globalConfigMap := mustConfigMap(dbConfigNode, configNodeNameDB)

		var databaseConfigMap map[string]any
		// try to get specific instance config
		if instanceConfig, ok := globalConfigMap[instanceName]; ok {
			databaseConfigMap = mustConfigMap(instanceConfig, fmt.Sprintf("%s.%s", configNodeNameDB, instanceName))
		} else if defaultConfig, ok := globalConfigMap["default"]; ok {
			// try to get default instance config
			databaseConfigMap = mustConfigMap(defaultConfig, configNodeNameDB+".default")
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

		// Prefer the instance logger configuration, then the scope logger.
		var loggerConfigMap map[string]any
		if loggerConfig, ok := databaseConfigMap[configNodeNameLogger]; ok {
			loggerConfigMap = mustConfigMap(loggerConfig, fmt.Sprintf("%s.%s.%s", configNodeNameDB, instanceName, configNodeNameLogger))
		} else if globalLoggerConfig, ok := configMap[configNodeNameLogger]; ok {
			loggerConfigMap = mustConfigMap(globalLoggerConfig, configNodeNameLogger)
		}

		// Attach a dedicated logger when configured.
		if len(loggerConfigMap) > 0 {
			dbLogger := mlog.New()
			if err := dbLogger.SetConfigWithMap(loggerConfigMap); err != nil {
				panic(merror.NewCodef(mcode.CodeInvalidConfiguration, "set db logger config failed: %v", err))
			}
			dbConfig.SetLogger(dbLogger)
		} else {
			// Otherwise, share the scope's default logger.
			dbConfig.SetLogger(s.Log())
		}

		// Create the connection pool only after configuration is complete.
		db, err := mdb.New(dbConfig)
		if err != nil {
			panic(err)
		}
		return db
	})

	return instance.(*mdb.DB)
}

// TryDB returns a database instance or an initialization error.
// Unlike DB, it does not panic when configuration or connection setup fails.
func TryDB(name ...string) (database *mdb.DB, err error) {
	return defaultScope.TryDB(name...)
}

// TryDB returns a scoped database instance or an initialization error.
func (s *Scope) TryDB(name ...string) (database *mdb.DB, err error) {
	defer recoverAsError(&err)
	return s.DB(name...), nil
}

// TryDBContext returns a context-bound database instance or an initialization error.
func TryDBContext(ctx context.Context, name ...string) (*mdb.DB, error) {
	return defaultScope.TryDBContext(ctx, name...)
}

// TryDBContext returns a context-bound database instance owned by the scope.
func (s *Scope) TryDBContext(ctx context.Context, name ...string) (*mdb.DB, error) {
	database, err := s.TryDB(name...)
	if err != nil {
		return nil, err
	}
	return database.WithContext(ctx), nil
}
