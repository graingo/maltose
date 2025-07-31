package internal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/graingo/maltose/os/mcfg/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test SearchMap ---

func TestSearchMap(t *testing.T) {
	source := map[string]any{
		"app": map[string]any{
			"name":    "maltose",
			"VERSION": "1.0",
		},
		"server": "localhost",
		"ports":  []int{80, 443},
	}

	t.Run("simple_path", func(t *testing.T) {
		result := internal.SearchMap(source, []string{"server"})
		assert.Equal(t, "localhost", result)
	})

	t.Run("nested_path", func(t *testing.T) {
		result := internal.SearchMap(source, []string{"app", "name"})
		assert.Equal(t, "maltose", result)
	})

	t.Run("case_insensitive_path", func(t *testing.T) {
		result := internal.SearchMap(source, []string{"App", "version"})
		assert.Equal(t, "1.0", result)
	})

	t.Run("path_not_found", func(t *testing.T) {
		result := internal.SearchMap(source, []string{"app", "host"})
		assert.Nil(t, result)
	})

	t.Run("path_through_non_map", func(t *testing.T) {
		result := internal.SearchMap(source, []string{"server", "port"})
		assert.Nil(t, result)
	})

	t.Run("empty_path", func(t *testing.T) {
		result := internal.SearchMap(source, []string{})
		assert.Equal(t, source, result)
	})

	t.Run("slice_value", func(t *testing.T) {
		result := internal.SearchMap(source, []string{"ports"})
		assert.Equal(t, []int{80, 443}, result)
	})
}

// --- Test DeepMergeMaps ---

func TestDeepMergeMaps(t *testing.T) {
	t.Run("add_new_keys", func(t *testing.T) {
		dest := map[string]any{"a": 1}
		src := map[string]any{"b": 2, "c": 3}
		internal.DeepMergeMaps(dest, src)
		expected := map[string]any{"a": 1, "b": 2, "c": 3}
		assert.Equal(t, expected, dest)
	})

	t.Run("overwrite_existing_key", func(t *testing.T) {
		dest := map[string]any{"a": 1}
		src := map[string]any{"a": "new"}
		internal.DeepMergeMaps(dest, src)
		expected := map[string]any{"a": "new"}
		assert.Equal(t, expected, dest)
	})

	t.Run("case_insensitive_overwrite", func(t *testing.T) {
		dest := map[string]any{"key": "old"}
		src := map[string]any{"Key": "new"}
		internal.DeepMergeMaps(dest, src)
		// The key from src should be preferred due to its casing.
		expected := map[string]any{"Key": "new"}
		assert.Equal(t, expected, dest)
	})

	t.Run("deep_merge_nested_maps", func(t *testing.T) {
		dest := map[string]any{
			"server": map[string]any{
				"host": "localhost",
			},
		}
		src := map[string]any{
			"server": map[string]any{
				"port": 8080,
			},
			"db": "postgres",
		}
		internal.DeepMergeMaps(dest, src)
		expected := map[string]any{
			"server": map[string]any{
				"host": "localhost",
				"port": 8080,
			},
			"db": "postgres",
		}
		assert.Equal(t, expected, dest)
	})

	t.Run("deep_merge_case_insensitive", func(t *testing.T) {
		dest := map[string]any{
			"Server": map[string]any{
				"Host": "localhost",
			},
		}
		src := map[string]any{
			"server": map[string]any{
				"port": 8080,
			},
		}
		internal.DeepMergeMaps(dest, src)
		expected := map[string]any{
			// Key casing from dest is preserved for the top-level map.
			"Server": map[string]any{
				"Host": "localhost",
				"port": 8080,
			},
		}
		assert.Equal(t, expected, dest)
	})

	t.Run("overwrite_map_with_value", func(t *testing.T) {
		dest := map[string]any{"a": map[string]any{"b": 1}}
		src := map[string]any{"a": "not a map"}
		internal.DeepMergeMaps(dest, src)
		expected := map[string]any{"a": "not a map"}
		assert.Equal(t, expected, dest)
	})
}

// --- Test FindCaseInsensitiveKey ---

func TestFindCaseInsensitiveKey(t *testing.T) {
	source := map[string]any{
		"KeyOne": "value1",
		"keyTwo": "value2",
	}

	t.Run("key_found_exact_case", func(t *testing.T) {
		key, found := internal.FindCaseInsensitiveKey(source, "KeyOne")
		assert.True(t, found)
		assert.Equal(t, "KeyOne", key)
	})

	t.Run("key_found_different_case", func(t *testing.T) {
		key, found := internal.FindCaseInsensitiveKey(source, "keyone")
		assert.True(t, found)
		assert.Equal(t, "KeyOne", key)
	})

	t.Run("key_not_found", func(t *testing.T) {
		_, found := internal.FindCaseInsensitiveKey(source, "KeyThree")
		assert.False(t, found)
	})
}

// --- Test SearchConfigFile ---

// setupTestFS creates a temporary directory structure for testing SearchConfigFile.
func setupTestFS(t *testing.T) (string, func()) {
	t.Helper()
	rootDir, err := os.MkdirTemp("", "search-config-test-*")
	require.NoError(t, err)

	configDir := filepath.Join(rootDir, "config")
	err = os.Mkdir(configDir, 0755)
	require.NoError(t, err)

	_, err = os.Create(filepath.Join(rootDir, "root-config.toml"))
	require.NoError(t, err)
	_, err = os.Create(filepath.Join(configDir, "my-app.json"))
	require.NoError(t, err)

	return rootDir, func() { os.RemoveAll(rootDir) }
}

func TestSearchConfigFile(t *testing.T) {
	rootDir, cleanup := setupTestFS(t)
	defer cleanup()

	originalWD, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(rootDir)
	require.NoError(t, err)
	defer os.Chdir(originalWD)

	t.Run("find_in_config_dir", func(t *testing.T) {
		path, found := internal.SearchConfigFile("my-app")
		assert.True(t, found)
		assert.Equal(t, filepath.Join("config", "my-app.json"), path)
	})

	t.Run("find_in_root_dir", func(t *testing.T) {
		path, found := internal.SearchConfigFile("root-config")
		assert.True(t, found)
		assert.Equal(t, "root-config.toml", path)
	})

	t.Run("find_with_extension", func(t *testing.T) {
		path, found := internal.SearchConfigFile("root-config.toml")
		assert.True(t, found)
		assert.Equal(t, "root-config.toml", path)
	})

	t.Run("file_not_found", func(t *testing.T) {
		_, found := internal.SearchConfigFile("non-existent")
		assert.False(t, found)
	})
}
