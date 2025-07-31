package apollo_test

import (
	"context"
	"os"
	"testing"

	"github.com/graingo/maltose/contrib/config/apollo"
	"github.com/graingo/maltose/frame/m"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ctx     = context.Background()
	ip      = os.Getenv("APOLLO_IP")
	appId   = os.Getenv("APOLLO_APP_ID")
	cluster = os.Getenv("APOLLO_CLUSTER")
)

func init() {
	if ip == "" {
		ip = "http://localhost:8080"
	}
	if appId == "" {
		appId = "SampleApp"
	}
	if cluster == "" {
		cluster = "default"
	}
}

func TestApollo(t *testing.T) {
	// Create adapter
	adapter, err := apollo.New(ctx, apollo.Config{
		AppID:     appId,
		IP:        ip,
		Cluster:   cluster,
		MustStart: true, // In test, we want to fail fast if connection fails.
	})
	require.NoError(t, err)
	require.NotNil(t, adapter)

	// Set configuration
	config := m.Config("apollo")
	config.SetAdapter(adapter)

	// Test availability
	// With MustStart: true, it should be available immediately.
	assert.True(t, config.Available(ctx))
	assert.False(t, config.Available(ctx, "non-exist"))

	// Test get configuration
	v, err := config.Get(ctx, `timeout`)
	assert.NoError(t, err)
	// The default value for 'timeout' in apollo-quick-start is 100
	assert.Equal(t, "100", v.String())

	// Test get another configuration
	v, err = config.Get(ctx, `server.address`)
	assert.NoError(t, err)
	// This key might not exist in default config, so we check for nil or empty.
	// In a real scenario, you'd populate this value in Apollo for the test.
	// For now, let's assume it's not there.
	assert.True(t, v == nil || v.String() == "")

	// Test get all configurations
	data, err := config.Data(ctx)
	assert.NoError(t, err)
	assert.Greater(t, len(data), 0)
	assert.Contains(t, data, "timeout")
}
