package mcfg_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/graingo/maltose/os/mcfg"
	"github.com/graingo/mconv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ctx = context.Background()
)

// setupTestConfigFile creates a temporary config file for testing.
func setupTestConfigFile(t *testing.T, dir, filename, content string) (string, func()) {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path, func() {
		os.Remove(path)
	}
}

// --- Initialization ---

func TestConfig_Initialization(t *testing.T) {
	// Point to the new testfile directory.
	testDir := "testfile"

	t.Run("new", func(t *testing.T) {
		// Temporarily change cwd to the testfile directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(testDir)
		require.NoError(t, err)
		defer os.Chdir(originalWd)

		c, err := mcfg.New()
		require.NoError(t, err)
		require.NotNil(t, c)
		assert.Equal(t, "test-app", c.GetString(ctx, "app.name"))
	})

	t.Run("new_with_adapter", func(t *testing.T) {
		adapter, err := mcfg.NewAdapterContent("app: {name: test-adapter}", "yaml")
		require.NoError(t, err)
		c := mcfg.NewWithAdapter(adapter)
		require.NotNil(t, c)
		val, err := c.Get(ctx, "app.name")
		require.NoError(t, err)
		assert.Equal(t, "test-adapter", val.String())
	})

	t.Run("instance", func(t *testing.T) {
		// Temporarily change cwd to the testfile directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(testDir)
		require.NoError(t, err)
		defer os.Chdir(originalWd)

		// Test default instance (should pick up config.yaml/yml/json/etc.)
		defaultInstance := mcfg.Instance()
		require.NotNil(t, defaultInstance)
		assert.Equal(t, "test-app", defaultInstance.GetString(ctx, "app.name"))

		// Test named instance from a different file type
		namedInstance := mcfg.Instance("config.toml")
		require.NotNil(t, namedInstance)
		assert.Equal(t, "test-app", namedInstance.GetString(ctx, "app.name"))

		// Test instance reuse
		sameInstance := mcfg.Instance("config.toml")
		assert.Same(t, namedInstance, sameInstance)
	})

	t.Run("instance_with_bad_config", func(t *testing.T) {
		// Create a temporary directory for the bad config file to avoid polluting the main test dir.
		tempDir, err := os.MkdirTemp("", "bad-config-test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		_, cleanup := setupTestConfigFile(t, tempDir, "bad-config.yml", `app: [name: "bad"`) // Malformed YAML
		defer cleanup()

		originalWd, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer os.Chdir(originalWd)

		assert.Panics(t, func() {
			mcfg.Instance("bad-config")
		})
	})

	t.Run("instance_file_not_found", func(t *testing.T) {
		// This should panic because SetFileName will return an error when the file is not found.
		assert.Panics(t, func() {
			mcfg.Instance("non-existent-config")
		})
	})
}

// --- Adapters ---

func TestAdapterFile(t *testing.T) {
	testDir := "testfile"

	t.Run("set_file_name_yaml", func(t *testing.T) {
		adapter, err := mcfg.NewAdapterFile()
		require.NoError(t, err)
		err = adapter.SetFile(filepath.Join(testDir, "config.yaml"))
		require.NoError(t, err)

		val, err := adapter.Get(ctx, "server.host")
		require.NoError(t, err)
		assert.Equal(t, "localhost", val)

		// Verify new values
		val, err = adapter.Get(ctx, "app.timeout")
		require.NoError(t, err)
		assert.Equal(t, "10s", val)

		val, err = adapter.Get(ctx, "owner.age.value")
		require.NoError(t, err)
		assert.Equal(t, 18, mconv.ToInt(val))
	})

	t.Run("set_file_name_json", func(t *testing.T) {
		adapter, err := mcfg.NewAdapterFile()
		require.NoError(t, err)
		err = adapter.SetFile(filepath.Join(testDir, "config.json"))
		require.NoError(t, err)

		val, err := adapter.Get(ctx, "server.port")
		require.NoError(t, err)
		assert.Equal(t, 8080, mconv.ToInt(val))

		// Verify new values
		val, err = adapter.Get(ctx, "app.parse2")
		require.NoError(t, err)
		assert.Equal(t, true, val)

		val, err = adapter.Get(ctx, "owner.age.unit")
		require.NoError(t, err)
		assert.Equal(t, "year", val)
	})

	t.Run("set_file_name_toml", func(t *testing.T) {
		adapter, err := mcfg.NewAdapterFile()
		require.NoError(t, err)
		err = adapter.SetFile(filepath.Join(testDir, "config.toml"))
		require.NoError(t, err)

		val, err := adapter.Get(ctx, "owner.name")
		require.NoError(t, err)
		assert.Equal(t, "gopher", val)

		// Verify new values
		val, err = adapter.Get(ctx, "app.parse")
		require.NoError(t, err)
		assert.Equal(t, int64(1), val) // TOML parser reads integers as int64

		val, err = adapter.Get(ctx, "owner.age.value")
		require.NoError(t, err)
		assert.Equal(t, int64(18), val)
	})

	t.Run("set_file_name_search_by_name", func(t *testing.T) {
		// Temporarily change cwd to the testfile directory so the adapter can find the file.
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(testDir)
		require.NoError(t, err)
		defer os.Chdir(originalWd)

		// Create a new adapter inside the correct directory context.
		// Note: NewAdapterFile will load the default "config" file automatically.
		adapter, err := mcfg.NewAdapterFile()
		require.NoError(t, err)

		// The default config should already be loaded, let's verify.
		val, err := adapter.Get(ctx, "app.name")
		require.NoError(t, err)
		assert.Equal(t, "test-app", val)

		// Now, let's explicitly set it again by its base name to test SetFileName's search logic.
		err = adapter.SetFile("config")
		require.NoError(t, err)
		val, err = adapter.Get(ctx, "app.version")
		require.NoError(t, err)
		assert.Equal(t, "1.0.0", val)
	})

	t.Run("available", func(t *testing.T) {
		adapter, err := mcfg.NewAdapterFile()
		require.NoError(t, err)
		err = adapter.SetFile(filepath.Join(testDir, "config.yaml"))
		require.NoError(t, err)

		assert.True(t, adapter.Available(ctx))
		assert.False(t, adapter.Available(ctx, "/path/to/nonexistent/file"))
	})
}

func TestAdapterContent(t *testing.T) {
	configContent := `
db:
  host: "localhost"
  port: 5432
`
	adapter, err := mcfg.NewAdapterContent(configContent, "yaml")
	require.NoError(t, err)

	t.Run("get", func(t *testing.T) {
		val, err := adapter.Get(ctx, "db.host")
		require.NoError(t, err)
		assert.Equal(t, "localhost", val)
	})

	t.Run("data", func(t *testing.T) {
		data, err := adapter.Data(ctx)
		require.NoError(t, err)
		require.Contains(t, data, "db")
		dbMap, ok := data["db"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "localhost", dbMap["host"])
	})

	t.Run("set_content", func(t *testing.T) {
		newContent := `db: {host: "remote", port: 5433}`
		err := adapter.SetContent(newContent, "yaml")
		require.NoError(t, err)

		val, err := adapter.Get(ctx, "db.host")
		require.NoError(t, err)
		assert.Equal(t, "remote", val)
	})

	t.Run("available", func(t *testing.T) {
		assert.True(t, adapter.Available(ctx))
	})
}

// --- Getters ---

func TestConfig_Getters(t *testing.T) {
	configContent := `
stringKey: "stringValue"
intKey: 123
boolKey: true
sliceKey: [1, 2, 3]
mapKey:
  nested: "value"
`
	adapter, err := mcfg.NewAdapterContent(configContent, "yaml")
	require.NoError(t, err)
	c := mcfg.NewWithAdapter(adapter)

	t.Run("get", func(t *testing.T) {
		val, err := c.Get(ctx, "stringKey")
		require.NoError(t, err)
		assert.Equal(t, "stringValue", val.String())

		val, err = c.Get(ctx, "nonexistent", "default")
		require.NoError(t, err)
		assert.Equal(t, "default", val.String())
	})

	t.Run("must_get", func(t *testing.T) {
		val := c.MustGet(ctx, "stringKey")
		assert.Equal(t, "stringValue", val.String())

		// MustGet does not panic on missing key if default is provided
		val = c.MustGet(ctx, "nonexistent", "default")
		assert.Equal(t, "default", val.String())
	})

	t.Run("get_string", func(t *testing.T) {
		assert.Equal(t, "stringValue", c.GetString(ctx, "stringKey"))
		assert.Equal(t, "default", c.GetString(ctx, "nonexistent", "default"))
		assert.Equal(t, "", c.GetString(ctx, "nonexistent"))
	})

	t.Run("get_int", func(t *testing.T) {
		assert.Equal(t, 123, c.GetInt(ctx, "intKey"))
		assert.Equal(t, 999, c.GetInt(ctx, "nonexistent", 999))
		assert.Equal(t, 0, c.GetInt(ctx, "nonexistent"))
	})

	t.Run("get_bool", func(t *testing.T) {
		assert.Equal(t, true, c.GetBool(ctx, "boolKey"))
		assert.Equal(t, false, c.GetBool(ctx, "nonexistent", false))
		assert.Equal(t, false, c.GetBool(ctx, "nonexistent"))
	})

	t.Run("get_slice", func(t *testing.T) {
		expected := []any{1, 2, 3}
		assert.Equal(t, expected, c.GetSlice(ctx, "sliceKey"))
	})

	t.Run("get_map", func(t *testing.T) {
		expected := map[string]any{"nested": "value"}
		assert.Equal(t, expected, c.GetMap(ctx, "mapKey"))
	})
}

// --- Struct Binding ---

func TestConfig_StructBinding(t *testing.T) {
	type ServerConfig struct {
		Host string `mconv:"host"`
		Port int    `mconv:"port"`
	}
	type AppConfig struct {
		Name   string       `mconv:"name"`
		Server ServerConfig `mconv:"server"`
	}

	configContent := `
name: "test-app"
server:
  host: "localhost"
  port: 8080
`
	adapter, err := mcfg.NewAdapterContent(configContent, "yaml")
	require.NoError(t, err)
	c := mcfg.NewWithAdapter(adapter)

	t.Run("bind_full_struct", func(t *testing.T) {
		var cfg AppConfig
		err := c.Struct(ctx, &cfg, "")
		require.NoError(t, err)
		assert.Equal(t, "test-app", cfg.Name)
		assert.Equal(t, "localhost", cfg.Server.Host)
		assert.Equal(t, 8080, cfg.Server.Port)
	})

	t.Run("bind_sub_struct", func(t *testing.T) {
		var cfg ServerConfig
		err := c.Struct(ctx, &cfg, "server")
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 8080, cfg.Port)
	})
}

// --- Hooks ---

func TestConfig_Hooks(t *testing.T) {
	t.Cleanup(mcfg.ClearHooks) // Ensure hooks are cleared after the test.

	configContent := `key: "original"`
	adapter, err := mcfg.NewAdapterContent(configContent, "yaml")
	require.NoError(t, err)

	hook := func(_ context.Context, data map[string]any) (map[string]any, error) {
		data["key"] = "hooked"
		data["newKey"] = "fromHook"
		return data, nil
	}
	mcfg.RegisterAfterLoadHook(hook)

	c := mcfg.NewWithAdapter(adapter)

	t.Run("data_with_hook", func(t *testing.T) {
		data, err := c.Data(ctx)
		require.NoError(t, err)
		assert.Equal(t, "hooked", data["key"])
		assert.Equal(t, "fromHook", data["newKey"])
	})
}

// --- Concurrency ---

func TestConfig_Concurrency(t *testing.T) {
	t.Cleanup(mcfg.ClearHooks) // Ensure no hooks from other tests interfere.

	content := `key: initial`
	adapter, err := mcfg.NewAdapterContent(content, "yaml")
	require.NoError(t, err)
	c := mcfg.NewWithAdapter(adapter)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Test concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			val := c.GetString(ctx, "key")
			assert.Equal(t, "initial", val)
		}()
	}
	wg.Wait()

	// Test concurrent read/write on adapter (through config)
	// NOTE: In a real app, changing adapter content while reading is not typical.
	// Here we use a content adapter which is safe for this test.
	contentAdapter, ok := c.GetAdapter().(*mcfg.AdapterContent)
	require.True(t, ok)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				// Concurrent read on the config object
				val := c.MustGet(ctx, "key", "").String()
				assert.Contains(t, []string{"initial", "value-1", "value-3", "value-5", "value-7", "value-9", "value-11", "value-13", "value-15", "value-17", "value-19", "value-21", "value-23", "value-25", "value-27", "value-29", "value-31", "value-33", "value-35", "value-37", "value-39", "value-41", "value-43", "value-45", "value-47", "value-49"}, val)
			} else {
				// Concurrent write on the underlying adapter
				newContent := fmt.Sprintf("key: value-%d", i)
				err := contentAdapter.SetContent(newContent, "yaml")
				assert.NoError(t, err)
				c.ClearCache(ctx) // Clear cache after content changes
			}
		}(i)
	}
	wg.Wait()
}
