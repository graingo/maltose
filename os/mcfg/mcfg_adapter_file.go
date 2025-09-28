package mcfg

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/graingo/maltose/os/mcfg/internal"
	"gopkg.in/yaml.v3"
)

// AdapterFile implements the Adapter interface for configuration in files.
type AdapterFile struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewAdapterFile creates a new file adapter and automatically loads the default config file if found.
func NewAdapterFile() (*AdapterFile, error) {
	a := &AdapterFile{
		data: make(map[string]any),
	}
	// Automatically search for and load the default config file.
	if path, found := internal.SearchConfigFile(DefaultConfigFileName); found {
		// Use the internal load method directly, as the instance is not yet shared.
		if err := a.load(path); err != nil {
			return nil, fmt.Errorf("error loading default config file at %s: %w", path, err)
		}
	}
	return a, nil
}

// SetFileName sets the configuration file by name or path and loads its content.
// Deprecated: Use SetFile instead.
func (c *AdapterFile) SetFileName(name string) error {
	return c.SetFile(name)
}

// SetFile sets the configuration file by name or path and loads its content.
// If the `name` is a base name without an extension (e.g., "config"), it will search for the file in default paths.
// Otherwise, it treats `name` as a full path.
func (c *AdapterFile) SetFile(name string) error {
	path := name
	// If `name` has no extension, try to find the config file with proper extension
	if filepath.Ext(name) == "" {
		// First check if the file exists as-is
		if _, err := os.Stat(name); os.IsNotExist(err) {
			// File doesn't exist as-is, try to find it with extensions
			if strings.Contains(name, string(os.PathSeparator)) {
				// Contains path separator, try adding extensions in the specified directory
				dir := filepath.Dir(name)
				baseName := filepath.Base(name)
				for _, ext := range []string{"yaml", "yml", "json", "toml"} {
					tryPath := filepath.Join(dir, baseName+"."+ext)
					if _, err := os.Stat(tryPath); err == nil {
						path = tryPath
						break
					}
				}
				// If still not found, fall back to global search with base name
				if path == name {
					if foundPath, found := internal.SearchConfigFile(baseName); found {
						path = foundPath
					} else {
						return fmt.Errorf("config file not found for name: %s", name)
					}
				}
			} else {
				// No path separator, use global search
				foundPath, found := internal.SearchConfigFile(name)
				if !found {
					return fmt.Errorf("config file not found for name: %s", name)
				}
				path = foundPath
			}
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.load(path)
}

// load reads, parses, and sets the configuration data from a given path.
// This is an internal method and assumes the caller handles locking.
func (c *AdapterFile) load(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found at path: %s", path)
		}
		return err
	}

	var data map[string]any
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	switch ext {
	case "yaml", "yml":
		err = yaml.Unmarshal(content, &data)
	case "json":
		// Use json.Unmarshal for direct parsing from []byte.
		err = json.Unmarshal(content, &data)
	case "toml":
		_, err = toml.Decode(string(content), &data)
	default:
		return fmt.Errorf("unsupported config file format: %s (extension: %s)", path, ext)
	}

	if err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	c.data = data
	return nil
}

// Get gets the configuration value for the specified key.
func (c *AdapterFile) Get(_ context.Context, pattern string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return internal.SearchMap(c.data, strings.Split(pattern, ".")), nil
}

// Data gets all configuration data.
func (c *AdapterFile) Data(_ context.Context) (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Return a copy to prevent external modification of the internal map.
	copyData := make(map[string]any, len(c.data))
	maps.Copy(copyData, c.data)
	return copyData, nil
}

// Available checks and returns whether the configuration service is available.
func (c *AdapterFile) Available(_ context.Context, resource ...string) bool {
	if len(resource) > 0 && resource[0] != "" {
		_, err := os.Stat(resource[0])
		return err == nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data) > 0
}

// MergeConfigMap merges a map into the existing configuration.
func (c *AdapterFile) MergeConfigMap(_ context.Context, data map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	Merge(c.data, data)
	return nil
}
