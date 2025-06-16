package mcfg

import (
	"context"
	"fmt"
	"sync"
)

// StatefulHook is an interface for hooks that need to maintain state across multiple calls.
// This is useful for caching results of expensive operations, like fetching remote configuration.
type StatefulHook interface {
	Hook(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)
}

type ConfigHookFunc func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)

var (
	// afterLoadHooks stores the hooks to be executed after configuration is loaded.
	afterLoadHooks     []ConfigHookFunc
	afterLoadHookMutex sync.RWMutex
)

// RegisterAfterLoadHook registers a hook to be executed after configuration is loaded.
// It accepts either a function with the signature `func(context.Context, map[string]interface{}) (map[string]interface{}, error)`
// or an implementation of the `StatefulHook` interface.
// Using a `StatefulHook` is the recommended way to handle expensive operations that should only run once (e.g., fetching remote config),
// as it allows caching within the hook's state.
func RegisterAfterLoadHook(hook interface{}) {
	afterLoadHookMutex.Lock()
	defer afterLoadHookMutex.Unlock()

	switch h := hook.(type) {
	case ConfigHookFunc:
		afterLoadHooks = append(afterLoadHooks, h)
	case func(context.Context, map[string]interface{}) (map[string]interface{}, error):
		afterLoadHooks = append(afterLoadHooks, h)
	case StatefulHook:
		afterLoadHooks = append(afterLoadHooks, h.Hook)
	default:
		panic(fmt.Sprintf("unsupported hook type: %T. Must be a ConfigHookFunc or a StatefulHook", h))
	}
}

// runAfterLoadHooks executes all registered after-load hooks in order.
func runAfterLoadHooks(ctx context.Context, data map[string]any) (map[string]any, error) {
	afterLoadHookMutex.RLock()
	hooks := make([]ConfigHookFunc, len(afterLoadHooks))
	copy(hooks, afterLoadHooks)
	afterLoadHookMutex.RUnlock()

	processedData := data
	var err error

	for _, hook := range hooks {
		processedData, err = hook(ctx, processedData)
		if err != nil {
			return nil, err
		}
	}

	return processedData, nil
}
