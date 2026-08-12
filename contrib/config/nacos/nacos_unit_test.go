package nacos

import (
	"context"
	"testing"

	"github.com/graingo/maltose/frame/m"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoUpdateRejectsInvalidContent(t *testing.T) {
	client := &Client{
		config: Config{ConfigParam: vo.ConfigParam{Type: "json"}},
		value:  m.NewVar(nil, true),
	}

	require.Error(t, client.doUpdate(`{"invalid":`))
	assert.True(t, client.value.IsNil())
}

func TestDoUpdateNormalizesTOML(t *testing.T) {
	client := &Client{
		config: Config{ConfigParam: vo.ConfigParam{DataId: "config.toml"}},
		value:  m.NewVar(nil, true),
	}

	require.NoError(t, client.doUpdate("[server]\naddress = ':9090'"))
	data, err := client.Data(context.Background())
	require.NoError(t, err)
	assert.Equal(t, ":9090", data["server"].(map[string]any)["address"])
}
