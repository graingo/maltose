package mlog

import (
	"reflect"

	"github.com/graingo/mconv"
)

type Config struct {
	// Level is the log level.
	Level Level `mapstructure:"level"`
	// Path is the log file path.
	Path string `mapstructure:"path"`
	// File is the log file name.
	File string `mapstructure:"file"`
	// TimeFormat is the log time format.
	TimeFormat string `mapstructure:"time_format"`
	// Format is the log format.
	Format string `mapstructure:"format"`
	// Stdout is the stdout print.
	Stdout bool `mapstructure:"stdout"`
	// AutoClean is the auto clean days.
	AutoClean int `mapstructure:"auto_clean"`
	// CtxKeys is the context keys to extract.
	CtxKeys []string `mapstructure:"ctx_keys"`
}

func defaultConfig() *Config {
	return &Config{
		Level:      defaultLevel,
		Path:       defaultPath,
		File:       defaultFile,
		TimeFormat: defaultTimeFormat,
		Format:     defaultFormat,
		Stdout:     true,
		CtxKeys:    []string{},
	}
}

// SetConfigWithMap sets the logger configuration using a map.
func (c *Config) SetConfigWithMap(configMap map[string]any) error {
	return mconv.StructE(configMap, c, stringToLevelHookFunc)
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
