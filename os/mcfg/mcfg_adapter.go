package mcfg

import "context"

// Adapter defines the configuration adapter interface.
type Adapter interface {
	// Get gets the configuration value for the specified key.
	Get(ctx context.Context, pattern string) (any, error)

	// Data gets all configuration data.
	Data(ctx context.Context) (map[string]any, error)

	// Available checks and returns whether the configuration service is available.
	// The optional `resource` parameter specifies certain configuration resources.
	Available(ctx context.Context, resource ...string) bool
}
