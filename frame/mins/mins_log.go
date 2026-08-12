package mins

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mlog"
)

const (
	configNodeNameLogger = "logger" // config node name for logger
)

// Log returns a logger instance from the default scope.
func Log(name ...string) *mlog.Logger {
	return defaultScope.Log(name...)
}

// Log returns a logger instance owned by the scope.
func (s *Scope) Log(name ...string) *mlog.Logger {
	var (
		ctx          = context.Background()
		instanceName = mlog.DefaultName
	)
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}
	instanceKey := fmt.Sprintf("%s.%s", frameCoreNameLogger, instanceName)

	instance := s.loggerInstances.GetOrSetFunc(instanceKey, func() any {
		logger := s.newLogger(instanceName)

		// It firstly searches configuration of the instance name.
		certainLoggerNodeName := fmt.Sprintf(`%s.%s`, configNodeNameLogger, instanceName)
		if v, err := s.Config().Get(ctx, certainLoggerNodeName); err != nil {
			panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `get logger config for instance "%s" failed: %v`, instanceName, err))
		} else if !v.IsNil() {
			if err := logger.SetConfigWithMap(v.Map()); err != nil {
				panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `set logger config for instance "%s" failed: %v`, instanceName, err))
			}
			return logger
		}

		// If the configuration for the instance name is not found,
		// it then searches the default configuration.
		if v, err := s.Config().Get(ctx, configNodeNameLogger); err != nil {
			panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `get default logger config failed: %v`, err))
		} else if !v.IsNil() {
			if err := logger.SetConfigWithMap(v.Map()); err != nil {
				panic(merror.NewCodef(mcode.CodeInvalidConfiguration, `set logger config for default failed: %v`, err))
			}
		}
		return logger
	})

	return instance.(*mlog.Logger)
}
