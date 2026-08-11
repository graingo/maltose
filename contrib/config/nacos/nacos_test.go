package nacos_test

import (
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/graingo/maltose/contrib/config/nacos"
	"github.com/graingo/maltose/os/mcfg"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ctx          = context.Background()
	nacosIPAddr  = "localhost"
	nacosPort    = uint64(8848)
	serverConfig constant.ServerConfig
	clientConfig constant.ClientConfig
)

func init() {
	if ip := os.Getenv("NACOS_IP_ADDR"); ip != "" {
		nacosIPAddr = ip
	}
	if portStr := os.Getenv("NACOS_PORT"); portStr != "" {
		if port, err := strconv.ParseUint(portStr, 10, 64); err == nil {
			nacosPort = port
		}
	}

	serverConfig = constant.ServerConfig{
		IpAddr: nacosIPAddr,
		Port:   nacosPort,
	}
	clientConfig = constant.ClientConfig{
		CacheDir:            "/tmp/nacos/cache",
		LogDir:              "/tmp/nacos/log",
		NotLoadCacheAtStart: true,
		LogLevel:            "warn",
	}
}

func setup(t *testing.T, dataID, group, content string) config_client.IConfigClient {
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: []constant.ServerConfig{serverConfig},
		},
	)
	require.NoError(t, err)

	_, err = configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
		Group:   group,
		Content: content,
		Type:    "toml",
	})
	require.NoError(t, err)

	// Wait for config to be published
	time.Sleep(2 * time.Second)

	return configClient
}

func teardown(t *testing.T, client config_client.IConfigClient, dataID, group string) {
	_, err := client.DeleteConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	require.NoError(t, err)
}

func TestNacos(t *testing.T) {
	dataID := "test-config-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	group := "test-group"
	initialContent := `
[server]
address = ":8080"
`
	client := setup(t, dataID, group, initialContent)
	defer teardown(t, client, dataID, group)

	type configChange struct {
		namespace string
		group     string
		dataID    string
		data      string
	}
	changeChan := make(chan configChange, 1)

	configParam := vo.ConfigParam{
		DataId: dataID,
		Group:  group,
		Type:   "toml",
	}

	// Create adapter with watch enabled
	adapter, err := nacos.New(ctx, nacos.Config{
		ServerConfigs: []constant.ServerConfig{serverConfig},
		ClientConfig:  clientConfig,
		ConfigParam:   configParam,
		Watch:         true,
		OnConfigChange: func(namespace, changedGroup, changedDataID, data string) {
			if !strings.Contains(data, "new-value") {
				return
			}
			select {
			case changeChan <- configChange{
				namespace: namespace,
				group:     changedGroup,
				dataID:    changedDataID,
				data:      data,
			}:
			default:
			}
		},
	})
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Create an isolated configuration instance backed by Nacos.
	cfg := mcfg.NewWithAdapter(adapter)

	// 1. Test initial configuration
	assert.True(t, cfg.Available(ctx))
	address, err := cfg.Get(ctx, "server.address")
	assert.NoError(t, err)
	assert.Equal(t, ":8080", address.String())

	allData, err := cfg.Data(ctx)
	assert.NoError(t, err)
	assert.Contains(t, allData, "server")

	// 2. Test configuration change
	newContent := `
[server]
address = ":9090"
new-key = "new-value"
`
	_, err = client.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
		Group:   group,
		Content: newContent,
		Type:    "toml",
	})
	require.NoError(t, err)

	select {
	case change := <-changeChan:
		assert.Equal(t, clientConfig.NamespaceId, change.namespace)
		assert.Equal(t, group, change.group)
		assert.Equal(t, dataID, change.dataID)
		assert.Contains(t, change.data, "new-value")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for configuration change")
	}

	// 3. Verify new configuration is loaded
	newAddress, err := cfg.Get(ctx, "server.address")
	assert.NoError(t, err)
	assert.Equal(t, ":9090", newAddress.String())

	newValue, err := cfg.Get(ctx, "new-key")
	assert.NoError(t, err)
	assert.Equal(t, "new-value", newValue.String())
}
