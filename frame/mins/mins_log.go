package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/os/mlog"
)

const (
	configNodeNameLogger = "logger" // config node name for logger
)

// Log returns a glog.Logger instance
// The parameter name is the instance name
func Log(name ...string) *mlog.Logger {
	var (
		ctx          = context.Background()
		instanceName = mlog.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}

	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameLogger, instanceName)
	v := globalInstances.GetOrSetFunc(instanceKey, func() any {
		// create logger instance
		logger := mlog.Instance(instanceName)
		// try to get logger config
		var configMap map[string]any
		// try to get logger config with certain name
		certainLoggerNodeName := fmt.Sprintf(`%s.%s`, configNodeNameLogger, instanceName)
		if v, _ := Config().Get(ctx, certainLoggerNodeName); v.IsNil() {
			configMap = v.Map()
		}
		// if certain config not exists, use global logger config
		if len(configMap) == 0 {
			if v, _ := Config().Get(ctx, configNodeNameLogger); !v.IsEmpty() {
				configMap = v.Map()
			}
		}
		// if config exists, set to logger instance
		if len(configMap) > 0 {
			if err := logger.SetConfigWithMap(configMap); err != nil {
				panic(err)
			}
		}
		return logger
	})

	return v.(*mlog.Logger)
}
