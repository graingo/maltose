package mins

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/graingo/maltose/os/mcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScopeBuildsNamedAndDefaultComponents(t *testing.T) {
	config := newConfigFromYAML(t, `
logger:
  service_name: root-service
  audit:
    service_name: audit-service
server:
  default:
    address: ":0"
  api:
    address: ":0"
database:
  default:
    type: sqlite
    dsn: "file:mins-default?mode=memory&cache=shared"
  reporting:
    type: sqlite
    dsn: "file:mins-reporting?mode=memory&cache=shared"
redis:
  default:
    address: "127.0.0.1:1"
  cache:
    address: "127.0.0.1:2"
`)
	scope := NewScope(config)

	assert.Same(t, config, scope.Config())
	assert.Equal(t, "root-service", scope.Log().GetConfig().ServiceName)
	assert.Equal(t, "audit-service", scope.Log("audit").GetConfig().ServiceName)
	assert.Same(t, scope.Server("api"), scope.Server("api"))
	assert.NotSame(t, scope.Server(), scope.Server("api"))

	database := scope.DB()
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	require.NoError(t, database.Ping(context.Background()))
	contextDB, err := scope.TryDBContext(context.WithValue(context.Background(), testContextKey{}, "value"))
	require.NoError(t, err)
	assert.NotSame(t, database, contextDB)
	assert.Same(t, database.GetLogger(), contextDB.GetLogger())

	reporting, err := scope.TryDB("reporting")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reporting.Close()) })
	assert.NotSame(t, database, reporting)

	redisClient := scope.Redis("cache")
	require.NotNil(t, redisClient)
	assert.Same(t, redisClient, scope.Redis("cache"))
	t.Cleanup(func() { require.NoError(t, redisClient.Close()) })
}

func TestScopeAcceptsFlatComponentConfiguration(t *testing.T) {
	databaseScope := NewScope(newConfigFromYAML(t, `
database:
  type: sqlite
  dsn: "file:mins-flat?mode=memory&cache=shared"
`))
	database, err := databaseScope.TryDB("custom")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })

	redisScope := NewScope(newConfigFromYAML(t, `
redis:
  address: "127.0.0.1:1"
`))
	client, err := redisScope.TryRedis("custom")
	require.NoError(t, err)
	require.NotNil(t, client)
	t.Cleanup(func() { require.NoError(t, client.Close()) })
}

func TestScopeReportsMalformedConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		yaml   string
		create func(*Scope) error
	}{
		{
			name: "database node is scalar",
			yaml: "database: invalid",
			create: func(scope *Scope) error {
				_, err := scope.TryDB()
				return err
			},
		},
		{
			name: "redis instance is scalar",
			yaml: "redis:\n  default: invalid",
			create: func(scope *Scope) error {
				_, err := scope.TryRedis()
				return err
			},
		},
		{
			name: "database config missing",
			yaml: "logger: {}",
			create: func(scope *Scope) error {
				_, err := scope.TryDB()
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.create(NewScope(newConfigFromYAML(t, tt.yaml)))
			assert.Error(t, err)
		})
	}
}

func TestScopeInternalGuards(t *testing.T) {
	assert.PanicsWithValue(t, "mins: scope must not be nil", func() {
		var scope *Scope
		scope.Config()
	})
	assert.Panics(t, func() { mustConfigMap("invalid", "node") })
	assert.Equal(t, map[string]any{"value": 1}, mustConfigMap(map[string]any{"value": 1}, "node"))

	originalErr := errors.New("original")
	assert.ErrorIs(t, recoveredPanic(originalErr), originalErr)
	assert.EqualError(t, recoveredPanic("string panic"), "framework instance initialization panicked: string panic")
	assert.NoError(t, recoveredPanic(nil))
	assert.Same(t, DefaultScope(), defaultScope)
}

func recoveredPanic(value any) (err error) {
	defer recoverAsError(&err)
	if value != nil {
		panic(value)
	}
	return nil
}

func newConfigFromYAML(t *testing.T, content string) *mcfg.Config {
	t.Helper()
	adapter, err := mcfg.NewAdapterContent(content, "yaml")
	require.NoError(t, err)
	config := mcfg.NewWithAdapter(adapter)
	require.True(t, config.Available(context.Background()), fmt.Sprintf("config must be available: %s", content))
	return config
}

type testContextKey struct{}
