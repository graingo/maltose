package mdb

import (
	"github.com/graingo/maltose/container/minstance"
)

const (
	DefaultName = "default"
)

var (
	instances = minstance.New()
	configs   = minstance.New()
)

// Instance returns the logger instance with the specified name.
func Instance(name ...string) (*DB, error) {
	key := DefaultName
	if len(name) > 0 && name[0] != "" {
		key = name[0]
	}

	v := instances.GetOrSetFunc(key, func() any {
		cfg := GetConfig(key)
		db, err := NewWithConfig(cfg)
		if err != nil {
			return err
		}
		return db
	})
	if err, ok := v.(error); ok {
		instances.Remove(key)
		return nil, err
	}
	return v.(*DB), nil
}

func SetConfig(name string, cfg *Config) {
	configs.Set(name, cfg)
}

func GetConfig(name string) *Config {
	if v := configs.Get(name); v != nil {
		return v.(*Config)
	}
	return nil
}
