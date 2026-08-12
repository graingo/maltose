package mdb

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

type testPlugin struct{ name string }

func (p testPlugin) Name() string            { return p.name }
func (testPlugin) Initialize(*gorm.DB) error { return nil }

func TestConfigMutationAndCloning(t *testing.T) {
	config := defaultConfig()
	require.NoError(t, config.SetConfigWithMap(map[string]any{
		"type": "sqlite", "dsn": "file:config-map?mode=memory&cache=shared",
		"max_open_connection": 4, "max_idle_time": "2s",
	}))
	assert.Equal(t, "sqlite", config.Type)
	assert.Equal(t, 4, config.MaxOpenConnection)
	assert.Equal(t, 2*time.Second, config.MaxIdleTime)

	config.SetLogger(nil)
	assert.NotNil(t, config.Logger)
	config.SetReplicas([]Config{{Host: "replica-a"}})
	config.AddReplica(Config{Host: "replica-b"})
	assert.Len(t, config.Replicas, 2)
	config.SetPlugins([]gorm.Plugin{testPlugin{name: "first"}})
	config.AddPlugin(testPlugin{name: "second"})
	assert.Len(t, config.Plugins, 2)

	cloned := cloneConfig(config)
	cloned.Replicas[0].Host = "changed"
	cloned.Plugins = nil
	assert.Equal(t, "replica-a", config.Replicas[0].Host)
	assert.Len(t, config.Plugins, 2)
	assert.NotNil(t, cloneConfig(nil))
}

func TestConfigRegistryAndInvalidInstance(t *testing.T) {
	const name = "additional-config"
	t.Cleanup(func() { RemoveConfig(name) })

	require.NoError(t, SetConfigByMap(map[string]any{
		"type": "sqlite", "dsn": "file:registry?mode=memory&cache=shared",
	}, name))
	config, ok := GetConfig(name)
	require.True(t, ok)
	assert.Equal(t, "sqlite", config.Type)

	config.Type = "mutated"
	stored, ok := GetConfig(name)
	require.True(t, ok)
	assert.Equal(t, "sqlite", stored.Type, "registry values must be returned as clones")
	require.NotNil(t, Instance(name))

	RemoveConfig(name)
	_, ok = GetConfig(name)
	assert.False(t, ok)
	assert.Nil(t, Instance(name))

	_, err := ConfigFromMap(map[string]any{"max_open_connection": []string{"invalid"}})
	assert.Error(t, err)
}

func TestTransactionsAndNilSafeMethods(t *testing.T) {
	db := newAdditionalTestDB(t)
	type record struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	require.NoError(t, db.AutoMigrate(&record{}))

	ctx := context.Background()
	require.NoError(t, db.Transact(ctx, func(tx *DB) error {
		return tx.Create(&record{Name: "committed"}).Error
	}))

	rollbackErr := errors.New("rollback")
	err := db.TransactWithOptions(ctx, &sql.TxOptions{}, func(tx *DB) error {
		require.NoError(t, tx.Create(&record{Name: "rolled-back"}).Error)
		return rollbackErr
	})
	require.ErrorIs(t, err, rollbackErr)

	var count int64
	require.NoError(t, db.Model(&record{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
	assert.NoError(t, (*DB)(nil).Close())
	assert.Nil(t, (&DB{}).GetLogger())
}

func TestGormLoggerLevelMethods(t *testing.T) {
	var output bytes.Buffer
	baseLogger := mlog.New(&mlog.Config{
		Writer: &output,
		Stdout: false,
		Level:  mlog.DebugLevel,
		Format: "json",
	})
	logger := NewGormLogger(baseLogger, WithLogLevel(gormlog.Info))
	changed := logger.LogMode(gormlog.Silent)
	assert.NotSame(t, logger, changed)
	assert.Equal(t, gormlog.Info, logger.gormLogLevel)

	ctx := context.Background()
	logger.Info(ctx, "info")
	logger.Warn(ctx, "warn")
	logger.Error(ctx, "error")
	assert.Contains(t, output.String(), `"msg":"info"`)
	assert.Contains(t, output.String(), `"msg":"warn"`)
	assert.Contains(t, output.String(), `"msg":"error"`)

	silent := changed.(*GormLogger)
	assert.Equal(t, gormlog.Silent, silent.gormLogLevel)
	outputBeforeSilentLogs := output.String()
	silent.Info(ctx, "ignored")
	silent.Warn(ctx, "ignored")
	silent.Error(ctx, "ignored")
	assert.Equal(t, outputBeforeSilentLogs, output.String())
}

func newAdditionalTestDB(t *testing.T) *DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := New(&Config{Type: "sqlite", DSN: dsn, Plugins: []gorm.Plugin{}})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return db
}
