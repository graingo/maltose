package mcfg

import (
	"context"
	"strings"
	"sync"

	"github.com/graingo/maltose/container/minstance"
	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mcfg/internal"
	"github.com/graingo/mconv"
)

var (
	instances = minstance.New()
)

// Config is a configuration management object.
type Config struct {
	adapter    Adapter
	cachedData *mvar.Var // Used to cache the data after hooks have been executed.
	mu         sync.RWMutex
}

const (
	// DefaultInstanceName is the default instance name.
	DefaultInstanceName = "default"
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
			panic(merror.Wrap(err, "create config instance failed"))
		}
		if instanceName != DefaultInstanceName {
			if err := adapterFile.SetFile(instanceName); err != nil {
				panic(merror.Wrapf(err, `set config file name for instance "%s" failed`, instanceName))
			}
		}
		return NewWithAdapter(adapterFile)
	}).(*Config)
}

// SetAdapter sets the configuration adapter.
func (c *Config) SetAdapter(adapter Adapter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.adapter = adapter
	c.cachedData = nil // Clear cache when adapter changes.
}

// GetAdapter gets the configuration adapter.
func (c *Config) GetAdapter() Adapter {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.adapter
}

// ClearCache clears the internal configuration cache.
// It should be called when the underlying configuration source has changed.
func (c *Config) ClearCache(_ context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cachedData = nil
}

// getValueByPattern gets the configuration value for the specified key.
// It uses a temporary viper instance to avoid concurrency issues on a shared instance.
func (c *Config) getValueByPattern(data map[string]any, pattern string) any {
	path := strings.Split(pattern, ".")
	return internal.SearchMap(data, path)
}

// Get gets the configuration value for the specified key.
// The optional `def` parameter is the default value. If the configuration value is empty, the default value is returned.
// If the configuration value is empty and no default value is provided, nil is returned.
func (c *Config) Get(ctx context.Context, pattern string, def ...any) (*mvar.Var, error) {
	data, err := c.Data(ctx)
	if err != nil {
		return nil, err
	}
	if data == nil {
		if len(def) > 0 {
			return mvar.New(def[0]), nil
		}
		return nil, nil
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
	// Use read lock to check for cached data.
	c.mu.RLock()
	if c.cachedData != nil {
		defer c.mu.RUnlock()
		return c.cachedData.Map(), nil
	}
	c.mu.RUnlock()

	// If no cache, use write lock to load and set data.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double check, as another goroutine might have populated it in the meantime.
	if c.cachedData != nil {
		return c.cachedData.Map(), nil
	}

	rawData, err := c.adapter.Data(ctx)
	if err != nil {
		return nil, err
	}

	if hooks.Count() > 0 {
		processedData, err := runAfterLoadHooks(ctx, rawData)
		if err != nil {
			return nil, err
		}
		c.cachedData = mvar.New(processedData)
		return processedData, nil
	}

	c.cachedData = mvar.New(rawData)
	return rawData, nil
}

// Available checks if the adapter is available.
// The optional `resource` parameter is the resource name. If the resource name is not empty, it checks if the resource is available.
func (c *Config) Available(ctx context.Context, resource ...string) bool {
	return c.adapter.Available(ctx, resource...)
}

// Struct unmarshals the configuration into a struct.
// The optional `pattern` parameter is the pattern to unmarshal the configuration into.
// If you want to specify the key name, you can use the `mconv` tag.
// It supports custom decoding hooks.
func (c *Config) Struct(ctx context.Context, v any, pattern string, hooks ...mconv.HookFunc) error {
	var (
		data map[string]any
		err  error
	)
	if pattern != "" {
		mvalue, err := c.Get(ctx, pattern)
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

	return mconv.ToStructE(data, v, hooks...)
}

// GetString gets the configuration value as a string.
func (c *Config) GetString(ctx context.Context, pattern string, def ...any) string {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	return val.String()
}

// GetInt gets the configuration value as an int.
func (c *Config) GetInt(ctx context.Context, pattern string, def ...any) int {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	return val.Int()
}

// GetBool gets the configuration value as a bool.
func (c *Config) GetBool(ctx context.Context, pattern string, def ...any) bool {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
	}
	return val.Bool()
}

// GetMap gets the configuration value as a map.
func (c *Config) GetMap(ctx context.Context, pattern string, def ...any) map[string]any {
	val, err := c.Get(ctx, pattern, def...)
	if err != nil {
		panic(err)
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
		return nil
	}
	return mconv.ToSlice(val.Val())
}
