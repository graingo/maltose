package mins

import (
	"context"
	"testing"

	"github.com/graingo/maltose/os/mcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopeIsolatesFrameworkInstances(t *testing.T) {
	scopeA := NewScope(newScopeTestConfig(t, "service-a"))
	scopeB := NewScope(newScopeTestConfig(t, "service-b"))

	loggerA := scopeA.Log()
	loggerB := scopeB.Log()
	assert.NotSame(t, loggerA, loggerB)
	assert.Same(t, loggerA, scopeA.Log())
	assert.Equal(t, "service-a", loggerA.GetConfig().ServiceName)
	assert.Equal(t, "service-b", loggerB.GetConfig().ServiceName)

	serverA := scopeA.Server()
	serverB := scopeB.Server()
	assert.NotSame(t, serverA, serverB)
	assert.Same(t, serverA, scopeA.Server())
}

func TestScopeTryMethodsUseScopedConfig(t *testing.T) {
	scope := NewScope(newScopeTestConfig(t, "service"))

	database, err := scope.TryDB("missing")
	assert.Nil(t, database)
	assert.Error(t, err)

	redisClient, err := scope.TryRedis("missing")
	assert.Nil(t, redisClient)
	assert.Error(t, err)
}

func TestNewScopeRequiresConfig(t *testing.T) {
	assert.PanicsWithValue(t, "mins: scope config must not be nil", func() {
		NewScope(nil)
	})
}

func newScopeTestConfig(t *testing.T, serviceName string) *mcfg.Config {
	t.Helper()
	adapter, err := mcfg.NewAdapterContent(`
logger:
  service_name: `+serviceName+`
server:
  address: ":0"
`, "yaml")
	require.NoError(t, err)
	config := mcfg.NewWithAdapter(adapter)
	require.True(t, config.Available(context.Background()))
	return config
}
