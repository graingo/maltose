package mcfg

import (
	"context"
	"os"
	"strings"

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

// NewAdapterFile creates a new file adapter.
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

// SetFileName sets the configuration file name.
func (c *AdapterFile) SetFileName(name string) {
	if name == "" {
		name = DefaultConfigFileName
	}
	// Remove the file extension if it exists.
	for _, ext := range supportedFileTypes {
		if strings.HasSuffix(name, "."+ext) {
			name = strings.TrimSuffix(name, "."+ext)
			break
		}
	}
	c.fileName = name
	c.v.SetConfigName(name)
	// read the config file again
	c.v.ReadInConfig()
}

// Get gets the configuration value.
func (c *AdapterFile) Get(ctx context.Context, pattern string) (any, error) {
	return c.v.Get(pattern), nil
}

// Data gets all configuration data.
func (c *AdapterFile) Data(ctx context.Context) (map[string]any, error) {
	return c.v.AllSettings(), nil
}

// Available checks and returns whether the configuration service is available.
// The optional `resource` parameter specifies certain configuration resources.
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

// MergeConfigMap merges a map into the existing configuration.
func (c *AdapterFile) MergeConfigMap(ctx context.Context, data map[string]any) error {
	return c.v.MergeConfigMap(data)
}

// IsExist checks if the file exists.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
