package nacos

import (
	"context"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/frame/m"
	"github.com/graingo/maltose/os/mcfg"
	nacosClients "github.com/nacos-group/nacos-sdk-go/v2/clients"
	nacosConfigClient "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
)

// Config is the configuration object for nacos client.
type Config struct {
	ServerConfigs  []constant.ServerConfig                     `binding:"required"` // See constant.ServerConfig
	ClientConfig   constant.ClientConfig                       `binding:"required"` // See constant.ClientConfig
	ConfigParam    vo.ConfigParam                              `binding:"required"` // See vo.ConfigParam
	Watch          bool                                        // Watch watches remote configuration updates, which updates local configuration in memory immediately when remote configuration changes.
	OnConfigChange func(namespace, group, dataId, data string) // Configure change callback function
}

// Client implements gcfg.Adapter implementing using nacos service.
type Client struct {
	config Config                          // Config object when created.
	client nacosConfigClient.IConfigClient // Nacos config client.
	value  *m.Var                          // Configmap content cached. It is json string.
}

// New creates and returns gcfg.Adapter implementing using nacos service.
func New(ctx context.Context, config Config) (adapte mcfg.Adapter, err error) {
	// Data validation.
	err = validator.New().Struct(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		config: config,
		value:  m.NewVar(nil, true),
	}

	client.client, err = nacosClients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": config.ServerConfigs,
		"clientConfig":  config.ClientConfig,
	})
	if err != nil {
		return nil, merror.Wrapf(err, `create nacos client failed with config: %+v`, config)
	}

	err = client.addWatcher()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Available checks and returns the backend configuration service is available.
// The optional parameter `resource` specifies certain configuration resource.
//
// Note that this function does not return error as it just does simply check for
// backend configuration service.
func (c *Client) Available(ctx context.Context, resource ...string) (ok bool) {
	if len(resource) == 0 && !c.value.IsNil() {
		return true
	}
	_, err := c.client.GetConfig(c.config.ConfigParam)
	return err == nil
}

// Get retrieves and returns value by specified `pattern` in current resource.
// Pattern like:
// "x.y.z" for map item.
// "x.0.y" for slice item.
func (c *Client) Get(ctx context.Context, pattern string) (value interface{}, err error) {
	if c.value.IsNil() {
		if err = c.updateLocalValue(); err != nil {
			return nil, err
		}
	}
	return gjson.Get(c.value.String(), pattern).Value(), nil
}

// Data retrieves and returns all configuration data in current resource as map.
// Note that this function may lead lots of memory usage if configuration data is too large,
// you can implement this function if necessary.
func (c *Client) Data(ctx context.Context) (data map[string]interface{}, err error) {
	if c.value.IsNil() {
		if err = c.updateLocalValue(); err != nil {
			return nil, err
		}
	}
	return gjson.Parse(c.value.String()).Value().(map[string]any), nil
}

func (c *Client) addWatcher() error {
	if !c.config.Watch {
		return nil
	}
	c.config.ConfigParam.OnChange = func(namespace, group, dataId, data string) {
		c.doUpdate(data)
		if c.config.OnConfigChange != nil {
			go c.config.OnConfigChange(namespace, group, dataId, data)
		}
	}

	if err := c.client.ListenConfig(c.config.ConfigParam); err != nil {
		return merror.Wrap(err, `watch config from namespace failed`)
	}

	return nil
}

func (c *Client) doUpdate(content string) (err error) {
	if !gjson.Valid(content) {
		return merror.Wrap(err, `parse config map item from nacos failed`)
	}
	c.value.Set(content)
	return nil
}

func (c *Client) updateLocalValue() (err error) {
	content, err := c.client.GetConfig(c.config.ConfigParam)
	if err != nil {
		return merror.Wrap(err, `retrieve config from nacos failed`)
	}

	return c.doUpdate(content)
}

func (c *Client) MergeConfigMap(ctx context.Context, data map[string]any) error {
	currentData, err := c.Data(ctx)
	if err != nil {
		return merror.Wrap(err, "failed to get current config")
	}

	// Use viper for deep merging
	v := viper.New()
	if err := v.MergeConfigMap(currentData); err != nil {
		return merror.Wrap(err, "failed to merge current config")
	}
	if err := v.MergeConfigMap(data); err != nil {
		return merror.Wrap(err, "failed to merge new data")
	}

	mergedData := v.AllSettings()
	mergedJSON, err := json.Marshal(mergedData)
	if err != nil {
		return merror.Wrap(err, "failed to marshal merged config")
	}

	c.doUpdate(string(mergedJSON))

	return nil
}
