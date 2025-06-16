package mcfg

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/container/minstance"
	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/errors/merror"
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
			_ = fmt.Errorf(`create config instance failed: %+v`, err)
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

// Data gets all configuration data.
func (c *Config) Data(ctx context.Context) (map[string]any, error) {
	rawData, err := c.adapter.Data(ctx)
	if err != nil {
		return nil, err
	}

	if len(afterLoadHooks) > 0 {
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

// MergeConfigMap merges a map into the existing configuration.
// This is useful for layering configurations.
func (c *Config) MergeConfigMap(ctx context.Context, data map[string]any) error {
	if c.adapter == nil {
		return merror.New(`config adapter is not set`)
	}
	return c.adapter.MergeConfigMap(ctx, data)
}
