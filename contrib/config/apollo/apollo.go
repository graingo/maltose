package apollo

import (
	"context"

	"github.com/apolloconfig/agollo/v4"
	apolloConfig "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/go-playground/validator/v10"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/frame/m"
	"github.com/graingo/maltose/os/mcfg"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Config is the configuration object for apollo client.
type Config struct {
	AppID             string `binding:"required"` // See apolloConfig.Config.
	IP                string `binding:"required"` // See apolloConfig.Config.
	Cluster           string `binding:"required"` // See apolloConfig.Config.
	NamespaceName     string // See apolloConfig.Config.
	IsBackupConfig    bool   // See apolloConfig.Config.
	BackupConfigPath  string // See apolloConfig.Config.
	Secret            string // See apolloConfig.Config.
	SyncServerTimeout int    // See apolloConfig.Config.
	MustStart         bool   // See apolloConfig.Config.
	Watch             bool   // Watch watches remote configuration updates, which updates local configuration in memory immediately when remote configuration changes.
}

// Client implements mcfg.Adapter implementing using apollo service.
type Client struct {
	config Config        // Config object when created.
	client agollo.Client // Apollo client.
	value  *m.Var        // Configmap content cached. It is json string.
}

// New creates and returns mcfg.Adapter implementing using apollo service.
func New(ctx context.Context, config Config) (adapter mcfg.Adapter, err error) {
	// Data validation.
	err = validator.New().Struct(config)
	if err != nil {
		return nil, err
	}

	if config.NamespaceName == "" {
		config.NamespaceName = storage.GetDefaultNamespace()
	}
	client := &Client{
		config: config,
		value:  m.NewVar(nil, true),
	}

	// Apollo client.
	client.client, err = agollo.StartWithConfig(func() (*apolloConfig.AppConfig, error) {
		return &apolloConfig.AppConfig{
			AppID:             config.AppID,
			Cluster:           config.Cluster,
			NamespaceName:     config.NamespaceName,
			IP:                config.IP,
			IsBackupConfig:    config.IsBackupConfig,
			BackupConfigPath:  config.BackupConfigPath,
			Secret:            config.Secret,
			SyncServerTimeout: config.SyncServerTimeout,
			MustStart:         config.MustStart,
		}, nil
	})
	if err != nil {
		return nil, merror.Wrapf(err, `create apollo client failed with config: %+v`, config)
	}
	if config.Watch {
		client.client.AddChangeListener(client)
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
	var namespace = c.config.NamespaceName
	if len(resource) > 0 {
		namespace = resource[0]
	}
	return c.client.GetConfig(namespace) != nil
}

// Get retrieves and returns value by specified `pattern` in current resource.
// Pattern like:
// "x.y.z" for map item.
// "x.0.y" for slice item.
func (c *Client) Get(ctx context.Context, pattern string) (value any, err error) {
	if c.value.IsNil() {
		if err = c.updateLocalValue(ctx); err != nil {
			return nil, err
		}
	}
	return gjson.Get(c.value.String(), pattern).Value(), nil
}

// Data retrieves and returns all configuration data in current resource as map.
// Note that this function may lead lots of memory usage if configuration data is too large,
// you can implement this function if necessary.
func (c *Client) Data(ctx context.Context) (data map[string]any, err error) {
	if c.value.IsNil() {
		if err = c.updateLocalValue(ctx); err != nil {
			return nil, err
		}
	}
	return gjson.Parse(c.value.String()).Value().(map[string]any), nil
}

// OnChange is called when config changes.
func (c *Client) OnChange(event *storage.ChangeEvent) {
	_ = c.updateLocalValue(context.Background())
}

// OnNewestChange is called when any config changes.
func (c *Client) OnNewestChange(event *storage.FullChangeEvent) {
	// Nothing to do.
}

func (c *Client) updateLocalValue(_ context.Context) (err error) {
	var s = ""
	cache := c.client.GetConfigCache(c.config.NamespaceName)
	cache.Range(func(key, value any) bool {
		s, err = sjson.Set(s, cast.ToString(key), value)
		return err == nil
	})
	cache.Clear()
	if err == nil {
		c.value.Set(s)
	}
	return
}
