package mdb

import (
	"context"

	"github.com/graingo/maltose/container/minstance"
	"github.com/graingo/maltose/internal/intlog"
	"github.com/spf13/viper"
)

const (
	// DefaultName is the default group name for db instance.
	DefaultName = "default"
)

var (
	// instances is a map for managing db instances.
	instances = minstance.New()
	// configs is a map for managing db configs.
	configs = minstance.New()
)

// Instance returns a db instance.
func Instance(name ...string) *DB {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}

	v := instances.GetOrSetFunc(key, func() any {
		if config, ok := GetConfig(key); ok {
			r, err := New(config)
			if err != nil {
				intlog.Errorf(context.TODO(), `new db instance failed: "%s"`, key)
				return nil
			}
			return r
		}
		return nil
	})
	if v != nil {
		return v.(*DB)
	}
	return nil
}

// SetConfig sets the db configuration with the specified name.
func SetConfig(name string, cfg *Config) {
	configs.Set(name, cfg)
}

// SetConfigByMap sets the db configuration with the specified name.
func SetConfigByMap(m map[string]any, name ...string) error {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	config, err := ConfigFromMap(m)
	if err != nil {
		return err
	}
	configs.Set(key, config)
	return nil
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

// GetConfig returns the db configuration with the specified name.
// If `name` is not passed, it returns configuration of the default name.
func GetConfig(name ...string) (config *Config, ok bool) {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	if v := configs.Get(key); v != nil {
		return v.(*Config), true
	}
	return &Config{}, false
}

// RemoveConfig removes the db configuration with the specified name.
func RemoveConfig(name ...string) {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}
	configs.Remove(key)

	intlog.Printf(context.TODO(), `db configuration "%s" removed`, key)
}
