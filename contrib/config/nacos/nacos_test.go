package nacos_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/graingo/maltose/contrib/config/nacos"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ctx          = context.Background()
	serverConfig = constant.ServerConfig{
		IpAddr: "localhost",
		Port:   8848,
	}
	clientConfig = constant.ClientConfig{
		CacheDir: "/tmp/nacos",
		LogDir:   "/tmp/nacos",
	}
	configParam = vo.ConfigParam{
		DataId: "config.toml",
		Group:  "test",
	}
	configPublishUrl = "http://localhost:8848/nacos/v2/cs/config?type=toml&namespaceId=public&group=test&dataId=config.toml"
)

func TestNacos(t *testing.T) {
	// Create adapter
	adapter, err := nacos.New(ctx, nacos.Config{
		ServerConfigs: []constant.ServerConfig{serverConfig},
		ClientConfig:  clientConfig,
		ConfigParam:   configParam,
	})
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Test availability
	assert.True(t, adapter.Available(ctx))

	// Test get configuration
	v, err := adapter.Get(ctx, `server.address`)
	assert.NoError(t, err)
	assert.Equal(t, ":8000", v)

	// Test get all configurations
	data, err := adapter.Data(ctx)
	assert.NoError(t, err)
	assert.Greater(t, len(data), 0)
}

func TestNacosOnConfigChangeFunc(t *testing.T) {
	configChanged := make(chan bool)

	adapter, err := nacos.New(ctx, nacos.Config{
		ServerConfigs: []constant.ServerConfig{serverConfig},
		ClientConfig:  clientConfig,
		ConfigParam:   configParam,
		Watch:         true,
		OnConfigChange: func(namespace, group, dataId, data string) {
			assert.Equal(t, "public", namespace)
			assert.Equal(t, "test", group)
			assert.Equal(t, "config.toml", dataId)
			configChanged <- true
		},
	})
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Test configuration changes
	testData := map[string]interface{}{
		"app": map[string]interface{}{
			"name": "test",
		},
	}

	// Publish new configuration
	content, err := json.Marshal(testData)
	require.NoError(t, err)

	// Send request using standard http package
	_, err = http.Post(configPublishUrl+"&content="+url.QueryEscape(string(content)),
		"application/json", nil)
	require.NoError(t, err)

	// Wait for configuration change callback
	select {
	case <-configChanged:
		// Configuration changed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("configuration change timeout")
	}

	// Cleanup configuration
	delete(testData, "app")
	content, err = json.Marshal(testData)
	require.NoError(t, err)

	_, err = http.Post(configPublishUrl+"&content="+url.QueryEscape(string(content)),
		"application/json", nil)
	require.NoError(t, err)
}
