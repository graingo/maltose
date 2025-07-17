package mlog

import (
	"reflect"

	"github.com/graingo/mconv"
)

type Config struct {
	// ServiceName is the service name.
	ServiceName string `mconv:"service_name"`
	// Level is the log level.
	Level Level `mconv:"level"`
	// TimeFormat is the log time format.
	TimeFormat string `mconv:"time_format"`
	// Format is the log format. Only support "json" and "text".
	Format string `mconv:"format"`
	// Caller controls whether the caller’s file and line number are included in logs.
	// If true, the caller’s file and line number will be added to the log entries.
	Caller bool `mconv:"caller"`
	// Development is the development mode.
	// If true, the logger will be in development mode.
	// It will print the error stack trace.
	Development bool `mconv:"development"`
	// Filepath is the log file path.
	// e.g., /var/log/app.log or /var/log/app.{YYYYmmdd}.log
	Filepath string `mconv:"filepath"`
	// MaxSize is the maximum size in megabytes of the log file before it gets rotated.
	// It is only applicable for 'size' rotation type.
	MaxSize int `mconv:"max_size"` // (MB)
	// MaxBackups is the maximum number of old log files to retain.
	// It is only applicable for 'size' rotation type.
	MaxBackups int `mconv:"max_backups"` // (files)
	// MaxAge is the maximum number of days to retain old log files.
	// It is applicable for both 'size' and 'date' rotation types.
	MaxAge int `mconv:"max_age"` // (days)
	// Stdout is the stdout print.
	Stdout bool `mconv:"stdout"`
	// CtxKeys is the context keys to extract.
	CtxKeys []string `mconv:"ctx_keys"`
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		ServiceName: "maltose",
		Level:       defaultLevel,
		TimeFormat:  defaultTimeFormat,
		Format:      defaultFormat,
		Caller:      false,
		Development: false,
		Filepath:    defaultFile,
		MaxSize:     100,
		MaxAge:      7,
		MaxBackups:  10,
		Stdout:      true,
		CtxKeys:     []string{},
	}
}

// SetConfigWithMap sets the logger configuration using a map.
func (c *Config) SetConfigWithMap(configMap map[string]any) error {
	return mconv.ToStructE(configMap, c, stringToLevelHookFunc)
}

// stringToLevelHookFunc is a hook function that converts a string to a Level.
func stringToLevelHookFunc(from reflect.Type, to reflect.Type, data any) (any, error) {
	if from.Kind() != reflect.String {
		return data, nil
	}
	if to != reflect.TypeOf(Level(0)) {
		return data, nil
	}

	levelStr := data.(string)
	level, err := ParseLevel(levelStr)
	if err != nil {
		return data, err
	}

	return level, nil
}
