package mcfg

import (
	"context"

	"github.com/graingo/maltose/container/minstance"
	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/mconv"
	"github.com/spf13/viper"
)

var (
	instances = minstance.New()
)

// Config is a configuration management object.
type Config struct {
	adapter Adapter
}

const (
	// DefaultInstanceName is the default instance name.
	DefaultInstanceName = "config"
	// DefaultConfigFileName is the default config file name.
	DefaultConfigFileName = "config"
)

// New creates a new configuration management object and uses the file adapter.
func New() (*Config, error) {
	adapterFile, err := NewAdapterFile()
	if err != nil {
		return nil, err
	}
	return &Config{
		adapter: adapterFile,
	}, nil
}

// NewWithAdapter creates a new configuration management object with an adapter.
func NewWithAdapter(adapter Adapter) *Config {
	return &Config{
		adapter: adapter,
	}
}

// Instance returns a Config instance with default settings.
// The `name` parameter is the instance name. Note that if a file named "name.yaml" exists in the config directory, it will be used as the default config file.
//
// Note: If a file named "name.yaml" exists in the config directory, it will be used as the default config file.
// If a file named "name.yaml" does not exist in the config directory, the default config file name "config" will be used.
func Instance(name ...string) *Config {
	var instanceName = DefaultInstanceName
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}

	return instances.GetOrSetFunc(instanceName, func() any {
		adapterFile, err := NewAdapterFile()
		if err != nil {
			_ = merror.Wrap(err, "create config instance failed")
			return nil
		}
		if instanceName != DefaultInstanceName {
			adapterFile.SetFileName(instanceName)
		}
		return NewWithAdapter(adapterFile)
	}).(*Config)
}

// SetAdapter sets the configuration adapter.
func (c *Config) SetAdapter(adapter Adapter) {
	c.adapter = adapter
}

// GetAdapter gets the configuration adapter.
func (c *Config) GetAdapter() Adapter {
	return c.adapter
}

// getValueByPattern gets the configuration value for the specified key.
func (c *Config) getValueByPattern(data map[string]any, pattern string) any {
	v := viper.New()
	v.MergeConfigMap(data)
	return v.Get(pattern)
}

// Get gets the configuration value for the specified key.
// The optional `def` parameter is the default value. If the configuration value is empty, the default value is returned.
// If the configuration value is empty and no default value is provided, nil is returned.
func (c *Config) Get(ctx context.Context, pattern string, def ...any) (*mvar.Var, error) {
	data, err := c.Data(ctx)
	if err != nil {
		return nil, err
	}

	value := c.getValueByPattern(data, pattern)
	if value != nil {
		return mvar.New(value), nil
	}

	if len(def) > 0 {
		return mvar.New(def[0]), nil
	}
	return nil, nil
}

// MustGet acts as function Get, but it panics if error occurs.
func (c *Config) MustGet(ctx context.Context, pattern string, def ...any) *mvar.Var {
	v, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	return v
}

// Data gets all configuration data.
func (c *Config) Data(ctx context.Context) (map[string]any, error) {
	rawData, err := c.adapter.Data(ctx)
	if err != nil {
		return nil, err
	}

	if hooks.Count() > 0 {
		processedData, err := runAfterLoadHooks(ctx, rawData)
		if err != nil {
			return nil, err
		}
		return processedData, nil
	}

	return rawData, nil
}

// Available checks if the adapter is available.
// The optional `resource` parameter is the resource name. If the resource name is not empty, it checks if the resource is available.
func (c *Config) Available(ctx context.Context, resource ...string) bool {
	return c.adapter.Available(ctx, resource...)
}

// Unmarshal unmarshals the configuration into a struct.
// The optional `pattern` parameter is the pattern to unmarshal the configuration into.
// If you want to specify the key name, you can use the `mapstructure` tag.
func (c *Config) Unmarshal(ctx context.Context, v any, pattern ...string) error {
	var (
		data map[string]any
		err  error
	)
	if len(pattern) > 0 {
		mvalue, err := c.Get(ctx, pattern[0])
		if err != nil {
			return err
		}
		if mvalue == nil || mvalue.IsNil() {
			return nil
		}
		data = mvalue.Map()
	} else {
		data, err = c.Data(ctx)
		if err != nil {
			return err
		}
	}
	if data == nil {
		return nil
	}

	return mconv.StructE(data, v)
}

// GetString gets the configuration value as a string.
func (c *Config) GetString(ctx context.Context, pattern string, def ...any) string {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	if val == nil {
		if len(def) > 0 {
			return mvar.New(def[0]).String()
		}
		return ""
	}
	return val.String()
}

// GetInt gets the configuration value as an int.
func (c *Config) GetInt(ctx context.Context, pattern string, def ...any) int {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	if val == nil {
		if len(def) > 0 {
			return mvar.New(def[0]).Int()
		}
		return 0
	}
	return val.Int()
}

// GetBool gets the configuration value as a bool.
func (c *Config) GetBool(ctx context.Context, pattern string, def ...any) bool {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	if val == nil {
		if len(def) > 0 {
			return mvar.New(def[0]).Bool()
		}
		return false
	}
	return val.Bool()
}

// GetMap gets the configuration value as a map.
func (c *Config) GetMap(ctx context.Context, pattern string, def ...any) map[string]any {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	if val == nil {
		if len(def) > 0 {
			return mvar.New(def[0]).Map()
		}
		return nil
	}
	return val.Map()
}

// GetSlice gets the configuration value as a slice.
func (c *Config) GetSlice(ctx context.Context, pattern string, def ...any) []any {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	if val == nil {
		if len(def) > 0 {
			return mconv.ToSlice(mvar.New(def[0]).Val())
		}
		return nil
	}
	return mconv.ToSlice(val.Val())
}
