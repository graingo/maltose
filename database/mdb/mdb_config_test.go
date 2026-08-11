package mdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMergeConfigPreservesDefaults(t *testing.T) {
	config := mergeConfig(&Config{DSN: "user:pass@tcp(localhost:3306)/app"})

	assert.Equal(t, "mysql", config.Type)
	assert.Equal(t, "3306", config.Port)
	assert.Equal(t, 10, config.MaxIdleConnection)
	assert.Equal(t, 100, config.MaxOpenConnection)
	assert.Equal(t, 10*time.Second, config.MaxIdleTime)
	assert.NotNil(t, config.Logger)
	assert.NotEmpty(t, config.Plugins)
}

func TestMergeConfigAppliesOverrides(t *testing.T) {
	config := mergeConfig(&Config{
		Type:              "sqlite",
		DSN:               "file:test.db",
		MaxIdleConnection: 3,
		MaxOpenConnection: 7,
		Plugins:           []gorm.Plugin{},
	})

	assert.Equal(t, "sqlite", config.Type)
	assert.Equal(t, "file:test.db", config.DSN)
	assert.Equal(t, 3, config.MaxIdleConnection)
	assert.Equal(t, 7, config.MaxOpenConnection)
	assert.Empty(t, config.Plugins)
}

func TestMergeReplicaConfigInheritsConnectionFields(t *testing.T) {
	primary := &Config{
		Type:     "mysql",
		Port:     "3306",
		User:     "app",
		Password: "secret",
		DBName:   "app_db",
	}
	replica := mergeReplicaConfig(primary, Config{Host: "replica.internal"})

	require.Equal(t, "replica.internal", replica.Host)
	assert.Equal(t, primary.Type, replica.Type)
	assert.Equal(t, primary.Port, replica.Port)
	assert.Equal(t, primary.User, replica.User)
	assert.Equal(t, primary.Password, replica.Password)
	assert.Equal(t, primary.DBName, replica.DBName)
}
