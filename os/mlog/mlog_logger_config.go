package mlog

import (
	"reflect"

	"github.com/graingo/mconv"
)

type Config struct {
	// Level is the log level.
	Level Level `mconv:"level"`
	// TimeFormat is the log time format.
	TimeFormat string `mconv:"time_format"`
	// Format is the log format.
	Format string `mconv:"format"`
	// Path is the log file path.
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

func defaultConfig() *Config {
	return &Config{
		Level:      defaultLevel,
		TimeFormat: defaultTimeFormat,
		Format:     defaultFormat,
		Filepath:   defaultFile,
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 10,
		Stdout:     true,
		CtxKeys:    []string{},
	}
}

// SetConfigWithMap sets the logger configuration using a map.
func (c *Config) SetConfigWithMap(configMap map[string]any) error {
	return mconv.ToStructE(configMap, c, stringToLevelHookFunc)
}

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
