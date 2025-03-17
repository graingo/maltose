package apollo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/graingo/maltose/contrib/config/apollo"
	"github.com/graingo/maltose/frame/m"
)

var (
	ctx     = context.Background()
	appId   = "SampleApp"
	cluster = "default"
	ip      = "http://localhost:8080"
)

func TestApollo(t *testing.T) {
	// Create adapter
	adapter, err := apollo.New(ctx, apollo.Config{
		AppID:   appId,
		IP:      ip,
		Cluster: cluster,
	})
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Set configuration
	config := m.Config("apollo")
	config.SetAdapter(adapter)

	// Test availability
	assert.True(t, config.Available(ctx))
	assert.False(t, config.Available(ctx, "non-exist"))

	// Test get configuration
	v, err := config.Get(ctx, `server.address`)
	assert.NoError(t, err)
	assert.Equal(t, ":8000", v.String())

	// Test get all configurations
	data, err := config.Data(ctx)
	assert.NoError(t, err)
	assert.Greater(t, len(data), 0)
}
