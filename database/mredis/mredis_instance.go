package mredis

import (
	"fmt"

	"github.com/graingo/maltose/container/minstance"
	"github.com/graingo/maltose/database/mredis/config"
	"github.com/graingo/maltose/internal/intlog"
)

const (
	// DefaultGroupName is the default group name for redis instance.
	DefaultGroupName = "default"
)

var (
	// instances is a map for managing redis instances.
	instances = minstance.New()
	// configs is a map for managing redis configs.
	configs = minstance.New()
)

// Instance returns a redis instance.
func Instance(name ...string) (*Redis, error) {
	group := DefaultGroupName
	if len(name) > 0 && name[0] != "" {
		group = name[0]
	}

	v := instances.GetOrSetFunc(group, func() any {
		cfg := GetConfig(group)
		if cfg == nil {
			return fmt.Errorf(`redis configuration "%s" is not found`, group)
		}
		db, err := NewWithConfig(cfg)
		if err != nil {
			return err
		}
		intlog.Printf(nil, `new redis instance created: "%s"`, group)
		return db
	})
	if err, ok := v.(error); ok {
		instances.Remove(group)
		return nil, err
	}
	return v.(*Redis), nil
}

func SetConfig(name string, cfg *config.Config) {
	configs.Set(name, cfg)
}

func GetConfig(name string) *config.Config {
	if v := configs.Get(name); v != nil {
		return v.(*config.Config)
	}
	return nil
}
