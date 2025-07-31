package mcfg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/graingo/maltose/os/mcfg/internal"
	"gopkg.in/yaml.v3"
)

// AdapterContent implements the Adapter interface for configuration in content.
type AdapterContent struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewAdapterContent creates a new adapter for configuration in content.
// The parameter `format` specifies the format of the content, e.g., "yaml", "json".
func NewAdapterContent(content string, format string) (*AdapterContent, error) {
	c := &AdapterContent{
		data: make(map[string]any),
	}
	if err := c.SetContent(content, format); err != nil {
		return nil, err
	}
	return c, nil
}

// SetContent sets the configuration content.
func (c *AdapterContent) SetContent(content string, format string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var data map[string]any
	var err error

	switch format {
	case "yaml", "yml":
		err = yaml.Unmarshal([]byte(content), &data)
	case "json":
		decoder := json.NewDecoder(bytes.NewReader([]byte(content)))
		decoder.UseNumber() // Use number to avoid converting numbers to float64
		err = decoder.Decode(&data)
	case "toml":
		_, err = toml.Decode(content, &data)
	default:
		return fmt.Errorf("unsupported config format: %s", format)
	}

	if err != nil {
		return err
	}
	c.data = data
	return nil
}

// Get gets the configuration value.
func (c *AdapterContent) Get(_ context.Context, pattern string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return internal.SearchMap(c.data, strings.Split(pattern, ".")), nil
}

// Data gets all configuration data.
func (c *AdapterContent) Data(_ context.Context) (map[string]any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Return a copy to prevent external modification of the internal map.
	copyData := make(map[string]any, len(c.data))
	maps.Copy(copyData, c.data)
	return copyData, nil
}

// Available checks and returns whether the configuration service is available.
// For the content adapter, it is always available if it was initialized successfully.
func (c *AdapterContent) Available(_ context.Context, _ ...string) bool {
	return c.data != nil
}

// MergeConfigMap merges a map into the existing configuration.
func (c *AdapterContent) MergeConfigMap(_ context.Context, data map[string]any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	Merge(c.data, data)
	return nil
}
