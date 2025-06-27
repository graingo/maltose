package mlogs

import (
	"reflect"

	"github.com/graingo/mconv"
)

const (
	defaultLevel  = InfoLevel
	defaultPath   = "./logs"
	defaultFile   = "maltose.log"
	defaultFormat = "text"
)

type Config struct {
	Level     Level    `mconv:"level"`
	Path      string   `mconv:"path"`
	File      string   `mconv:"file"`
	Format    string   `mconv:"format"` // "json" or "text"
	Stdout    bool     `mconv:"stdout"`
	AutoClean int      `mconv:"auto_clean"`
	CtxKeys   []string `mconv:"ctx_keys"`
	Hooks     []Hook
}

func defaultConfig() *Config {
	return &Config{
		Level:   defaultLevel,
		Path:    defaultPath,
		File:    defaultFile,
		Format:  defaultFormat,
		Stdout:  true,
		CtxKeys: []string{},
	}
}

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
