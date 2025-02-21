package gcfg

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/mingzaily/maltose/container/gvar"
	"github.com/spf13/viper"
)

const (
	// 默认实例名称
	DefaultInstanceName = "config"
	// 默认配置文件名
	DefaultConfigFileName = "config"
)

var (
	instances = make(map[string]*Config)
	mu        sync.RWMutex
)

type Config struct {
	v         *viper.Viper
	configDir string // 配置文件目录
	fileName  string // 配置文件名(不含扩展名)
	fileLock  sync.RWMutex
}

// Instance 返回指定名称的配置实例
func Instance(name ...string) *Config {
	instanceName := DefaultInstanceName
	if len(name) > 0 && name[0] != "" {
		instanceName = name[0]
	}

	mu.RLock()
	if ins, ok := instances[instanceName]; ok {
		mu.RUnlock()
		return ins
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// 创建新实例
	c := &Config{
		v:         viper.New(),
		configDir: "./config",
		fileName:  DefaultConfigFileName,
	}

	// 设置默认配置路径和名称
	c.SetPath("./config")
	c.SetFileName(DefaultConfigFileName)

	instances[instanceName] = c
	return c
}

// SetPath 设置配置文件搜索路径，支持链式调用
func (c *Config) SetPath(path string) *Config {
	c.fileLock.Lock()
	defer c.fileLock.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return c
	}

	c.configDir = absPath
	c.v.AddConfigPath(absPath)
	return c
}

// SetFileName 设置配置文件名(不含扩展名)，支持链式调用
func (c *Config) SetFileName(name string) *Config {
	c.fileLock.Lock()
	defer c.fileLock.Unlock()

	c.fileName = name
	c.v.SetConfigName(name)
	return c
}

// WithPath 同时设置路径和文件名，支持链式调用
func (c *Config) WithPath(path, fileName string) *Config {
	return c.SetPath(path).SetFileName(fileName)
}

// Reload 重新加载配置文件
func (c *Config) Reload() error {
	c.fileLock.Lock()
	defer c.fileLock.Unlock()

	// 重新创建 viper 实例
	c.v = viper.New()
	c.v.SetConfigName(c.fileName)
	c.v.AddConfigPath(c.configDir)

	// 支持的配置文件类型
	c.v.SetConfigType("yaml") // 默认支持 yaml
	c.v.SetConfigType("json") // 支持 json
	c.v.SetConfigType("toml") // 支持 toml

	// 读取配置文件
	if err := c.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	return nil
}

// Get 获取配置值
func (c *Config) Get(ctx context.Context, pattern string) *gvar.Var {
	c.fileLock.RLock()
	defer c.fileLock.RUnlock()
	return gvar.New(c.v.Get(pattern))
}

// Data 获取所有配置数据
func (c *Config) Data(ctx context.Context) (map[string]interface{}, error) {
	c.fileLock.RLock()
	defer c.fileLock.RUnlock()
	return c.v.AllSettings(), nil
}
