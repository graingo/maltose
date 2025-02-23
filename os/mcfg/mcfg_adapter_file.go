package mcfg

import (
	"context"
	"os"

	"github.com/spf13/viper"
)

var (
	supportedFileTypes = []string{"yaml", "yml", "json", "toml", "properties", "ini"}
	defaultConfigDir   = []string{"/", "config/", "config", "/config", "/config/", "./config"}
)

type AdapterFile struct {
	v        *viper.Viper
	fileName string
}

// NewAdapterFile 创建一个新的文件适配器
func NewAdapterFile() (*AdapterFile, error) {
	v := viper.New()
	v.SetConfigName(DefaultConfigFileName)

	for _, dir := range defaultConfigDir {
		v.AddConfigPath(dir)
	}

	v.ReadInConfig()
	return &AdapterFile{
		v:        v,
		fileName: DefaultConfigFileName,
	}, nil
}

// SetFileName 设置配置文件名
func (c *AdapterFile) SetFileName(name string) {
	c.fileName = name
	c.v.SetConfigName(name)
	// 重新读取配置文件
	c.v.ReadInConfig()
}

// Get 获取配置值
func (c *AdapterFile) Get(ctx context.Context, pattern string) (any, error) {
	return c.v.Get(pattern), nil
}

// Data 获取所有配置数据
func (c *AdapterFile) Data(ctx context.Context) (map[string]any, error) {
	return c.v.AllSettings(), nil
}

// Available 检查和后端配置服务是否可用。
// 可选参数 `resource` 指定某些配置资源。
func (c *AdapterFile) Available(ctx context.Context, resource ...string) bool {
	checkFileName := c.fileName
	if len(resource) > 0 && resource[0] != "" {
		checkFileName = resource[0]
	}

	for _, dir := range defaultConfigDir {
		for _, fileType := range supportedFileTypes {
			path := dir + checkFileName + "." + fileType
			if IsExist(path) {
				return true
			}
		}
	}
	return false
}

// IsExist 检查文件是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
