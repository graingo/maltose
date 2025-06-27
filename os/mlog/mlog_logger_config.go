package mlog

import (
	"reflect"

	"github.com/graingo/mconv"
)

type Config struct {
	// Level is the log level.
	Level Level `mconv:"level"`
	// Path is the log file path.
	Path string `mconv:"path"`
	// File is the log file name.
	File string `mconv:"file"`
	// TimeFormat is the log time format.
	TimeFormat string `mconv:"time_format"`
	// Format is the log format.
	Format string `mconv:"format"`
	// Stdout is the stdout print.
	Stdout bool `mconv:"stdout"`
	// AutoClean is the auto clean days.
	AutoClean int `mconv:"auto_clean"`
	// CtxKeys is the context keys to extract.
	CtxKeys []string `mconv:"ctx_keys"`
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
