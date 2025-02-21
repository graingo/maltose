package gcfg_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mingzaily/maltose/os/gcfg"
	"github.com/stretchr/testify/assert"
)

var (
	ctx           = context.Background()
	fixtureConfig = `
app:
  name: "test-app"
  version: "1.0.0"
  debug: true
  port: 8080
database:
  host: "localhost"
  port: 5432
  settings:
    timeout: 30
    maxConn: 100
`
	fixtureConfigTest = `
app:
  name: "test-app-test"
  version: "2.0.0"
`
)

func TestBasicOperations(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(fixtureConfig), 0644)
	assert.NoError(t, err)

	// 初始化配置
	c := gcfg.Instance()
	err = c.SetPath(tmpDir)
	assert.NoError(t, err)
	err = c.Reload()
	assert.NoError(t, err)

	// 测试基本类型
	assert.Equal(t, "test-app", c.Get(ctx, "app.name").String())
	assert.Equal(t, "1.0.0", c.Get(ctx, "app.version").String())
	assert.Equal(t, true, c.Get(ctx, "app.debug").Bool())
	assert.Equal(t, 8080, c.Get(ctx, "app.port").Int())

	// 测试嵌套结构
	assert.Equal(t, "localhost", c.Get(ctx, "database.host").String())
	assert.Equal(t, 5432, c.Get(ctx, "database.port").Int())
	assert.Equal(t, 30, c.Get(ctx, "database.settings.timeout").Int())
	assert.Equal(t, 100, c.Get(ctx, "database.settings.maxConn").Int())

	// 测试不存在的键
	assert.Equal(t, "", c.Get(ctx, "not.exists").String())
	assert.Equal(t, 0, c.Get(ctx, "not.exists").Int())
	assert.Equal(t, false, c.Get(ctx, "not.exists").Bool())

	// 测试获取全部配置
	data, err := c.Data(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Contains(t, data, "app")
	assert.Contains(t, data, "database")
}

func TestMultipleInstances(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建两个不同的配置文件
	err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(fixtureConfig), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "config.test.yaml"), []byte(fixtureConfigTest), 0644)
	assert.NoError(t, err)

	// 测试默认实例
	c1 := gcfg.Instance()
	err = c1.SetPath(tmpDir)
	assert.NoError(t, err)
	err = c1.Reload()
	assert.NoError(t, err)

	// 测试命名实例
	c2 := gcfg.Instance("test")
	err = c2.WithPath(tmpDir, "config.test").Reload()
	assert.NoError(t, err)

	// 验证不同实例读取不同配置
	assert.Equal(t, "test-app", c1.Get(ctx, "app.name").String())
	assert.Equal(t, "test-app-test", c2.Get(ctx, "app.name").String())
	assert.Equal(t, "1.0.0", c1.Get(ctx, "app.version").String())
	assert.Equal(t, "2.0.0", c2.Get(ctx, "app.version").String())
}

func TestReloadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// 写入初始配置
	err := os.WriteFile(configPath, []byte(fixtureConfig), 0644)
	assert.NoError(t, err)

	c := gcfg.Instance("reload")
	err = c.WithPath(tmpDir, "config").Reload()
	assert.NoError(t, err)

	assert.Equal(t, "test-app", c.Get(ctx, "app.name").String())

	// 更新配置文件
	err = os.WriteFile(configPath, []byte(fixtureConfigTest), 0644)
	assert.NoError(t, err)

	// 重新加载配置
	err = c.Reload()
	assert.NoError(t, err)

	assert.Equal(t, "test-app-test", c.Get(ctx, "app.name").String())
}

func TestFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	c := gcfg.Instance("not-found")

	err := c.WithPath(tmpDir, "not-exists").Reload()
	assert.NoError(t, err) // 配置文件不存在不应该返回错误

	// 不存在的配置应该返回零值
	assert.Equal(t, "", c.Get(ctx, "any.key").String())
	assert.Equal(t, 0, c.Get(ctx, "any.key").Int())
	assert.Equal(t, false, c.Get(ctx, "any.key").Bool())
}
