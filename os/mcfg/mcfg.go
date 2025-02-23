package mcfg

import (
	"context"
	"fmt"

	"github.com/mingzaily/maltose/container/minstance"
	"github.com/mingzaily/maltose/container/mvar"
)

var (
	localInstances = minstance.New()
)

// Config 是配置管理对象
type Config struct {
	adapter Adapter
}

const (
	// DefaultInstanceName 默认实例名称
	DefaultInstanceName = "config"
	// DefaultConfigFileName 默认配置文件名
	DefaultConfigFileName = "config"
)

// New 创建一个新的配置管理对象，并使用文件适配器
func New() (*Config, error) {
	adapterFile, err := NewAdapterFile()
	if err != nil {
		return nil, err
	}
	return &Config{
		adapter: adapterFile,
	}, nil
}

// NewWithAdapter 创建一个新的配置管理对象
func NewWithAdapter(adapter Adapter) *Config {
	return &Config{
		adapter: adapter,
	}
}

// Instance 返回一个具有默认设置的 Config 实例
// 参数 `name` 是实例的名称。但需要注意的是，如果配置目录中存在文件 "name.yaml"，则将其设置为默认配置文件
//
// 注意：如果配置目录中存在文件 "name.yaml"，则将其设置为默认配置文件
// 如果配置目录中不存在文件 "name.yaml"，则使用默认配置文件名 "config"
func Instance(name ...string) *Config {
	var instanceName = DefaultInstanceName
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}

	v := localInstances.GetOrSetFunc(instanceName, func() any {
		adapterFile, err := NewAdapterFile()
		if err != nil {
			_ = fmt.Errorf(`create config instance failed: %+v`, err)
			return nil
		}
		if instanceName != DefaultInstanceName {
			adapterFile.SetFileName(instanceName)
		}
		return NewWithAdapter(adapterFile)
	})

	return v.(*Config)
}

// SetAdapter 设置配置适配器
func (c *Config) SetAdapter(adapter Adapter) {
	c.adapter = adapter
}

// GetAdapter 获取配置适配器
func (c *Config) GetAdapter() Adapter {
	return c.adapter
}

// Get 获取指定键的配置值
// 可选参数 `def` 是默认值，如果配置值为空，则返回默认值
// 如果配置值为空，并且没有提供默认值，则返回 nil
func (c *Config) Get(ctx context.Context, pattern string, def ...any) (*mvar.Var, error) {
	var (
		err   error
		value any
	)
	value, err = c.adapter.Get(ctx, pattern)
	if err != nil {
		return nil, err
	}
	if value == nil {
		if len(def) > 0 {
			return mvar.New(def[0]), nil
		}
		return nil, nil
	}
	return mvar.New(value), nil
}

// Data 获取所有配置数据
func (c *Config) Data(ctx context.Context) (map[string]any, error) {
	return c.adapter.Data(ctx)
}

// Available 检查适配器是否可用
// 可选参数 `resource` 是资源名称，如果资源名称不为空，则检查资源是否可用
func (c *Config) Available(ctx context.Context, resource ...string) bool {
	return c.adapter.Available(ctx, resource...)
}
