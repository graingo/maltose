package nacos_test

import (
	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/graingo/maltose/contrib/config/nacos"
	"github.com/graingo/maltose/frame/m"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ctx          = context.Background()
	nacosIpAddr  = "localhost"
	nacosPort    = uint64(8848)
	serverConfig constant.ServerConfig
	clientConfig constant.ClientConfig
)

func init() {
	if ip := os.Getenv("NACOS_IP_ADDR"); ip != "" {
		nacosIpAddr = ip
	}
	if portStr := os.Getenv("NACOS_PORT"); portStr != "" {
		if port, err := strconv.ParseUint(portStr, 10, 64); err == nil {
			nacosPort = port
		}
	}

	serverConfig = constant.ServerConfig{
		IpAddr: nacosIpAddr,
		Port:   nacosPort,
	}
	clientConfig = constant.ClientConfig{
		CacheDir:            "/tmp/nacos/cache",
		LogDir:              "/tmp/nacos/log",
		NotLoadCacheAtStart: true,
		LogLevel:            "warn",
	}
}

func setup(t *testing.T, dataId, group, content string) config_client.IConfigClient {
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: []constant.ServerConfig{serverConfig},
		},
	)
	require.NoError(t, err)

	_, err = configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataId,
		Group:   group,
		Content: content,
		Type:    "toml",
	})
	require.NoError(t, err)

	// Wait for config to be published
	time.Sleep(2 * time.Second)

	return configClient
}

func teardown(t *testing.T, client config_client.IConfigClient, dataId, group string) {
	_, err := client.DeleteConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	require.NoError(t, err)
}

func TestNacos(t *testing.T) {
	dataId := "test-config-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	group := "test-group"
	initialContent := `
[server]
address = ":8080"
`
	client := setup(t, dataId, group, initialContent)
	defer teardown(t, client, dataId, group)

	var wg sync.WaitGroup
	wg.Add(1)

	configParam := vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	}

	// Create adapter with watch enabled
	adapter, err := nacos.New(ctx, nacos.Config{
		ServerConfigs: []constant.ServerConfig{serverConfig},
		ClientConfig:  clientConfig,
		ConfigParam:   configParam,
		Watch:         true,
		OnConfigChange: func(namespace, group, dataId, data string) {
			defer wg.Done()
			assert.Equal(t, "public", namespace)
			assert.Equal(t, group, group)
			assert.Equal(t, dataId, dataId)
			assert.Contains(t, data, "new-value")
		},
	})
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Use adapter in config instance
	cfg := m.Config("nacos")
	cfg.SetAdapter(adapter)

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
		DataId:  dataId,
		Group:   group,
		Content: newContent,
		Type:    "toml",
	})
	require.NoError(t, err)

	// Wait for the configuration change to be processed
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		// success
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
