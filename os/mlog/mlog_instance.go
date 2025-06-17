package mlog

import (
	"github.com/graingo/maltose/container/minstance"
	"github.com/spf13/viper"
)

const (
	DefaultName = "default"
)

var (
	instances = minstance.New()
)

// Instance returns the logger instance with the specified name.
func Instance(name ...string) *Logger {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}

	return instances.GetOrSetFunc(key, func() any {
		return New()
	}).(*Logger)
}

// ConfigFromMap parses and returns config from given map.
func ConfigFromMap(m map[string]any) (config *Config, err error) {
	v := viper.New()
	v.MergeConfigMap(m)
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}
	return config, nil
}
