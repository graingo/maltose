package mins

import (
	"testing"

	"github.com/graingo/maltose/os/mcfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTryInstancesReturnConfigurationErrors(t *testing.T) {
	config := Config()
	originalAdapter := config.GetAdapter()
	adapter, err := mcfg.NewAdapterContent("logger: {}", "yaml")
	require.NoError(t, err)
	config.SetAdapter(adapter)
	t.Cleanup(func() {
		config.SetAdapter(originalAdapter)
	})

	database, err := TryDB("missing-database-test")
	assert.Nil(t, database)
	assert.Error(t, err)

	redisClient, err := TryRedis("missing-redis-test")
	assert.Nil(t, redisClient)
	assert.Error(t, err)
}
