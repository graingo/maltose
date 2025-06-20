package mcfg

import (
	"context"
	"fmt"

	"github.com/graingo/maltose/container/minstance"
)

// StatefulHook is an interface for hooks that need to maintain state across multiple calls.
// This is useful for caching results of expensive operations, like fetching remote configuration.
type StatefulHook interface {
	Hook(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)
}

type ConfigHookFunc func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)

var (
	// hooks stores the hooks to be executed after configuration is loaded,
	// using a concurrent-safe instance manager.
	hooks = minstance.New()
)

// RegisterAfterLoadHook registers a hook to be executed after configuration is loaded.
// It accepts either a function with the signature `func(context.Context, map[string]interface{}) (map[string]interface{}, error)`
// or an implementation of the `StatefulHook` interface.
// Using a `StatefulHook` is the recommended way to handle expensive operations that should only run once (e.g., fetching remote config),
// as it allows caching within the hook's state.
// Each hook is stored with a unique key to prevent duplicate registrations.
func RegisterAfterLoadHook(hook interface{}) {
	var hookFunc ConfigHookFunc
	switch h := hook.(type) {
	case ConfigHookFunc:
		hookFunc = h
	case func(context.Context, map[string]interface{}) (map[string]interface{}, error):
		hookFunc = h
	case StatefulHook:
		hookFunc = h.Hook
	default:
		panic(fmt.Sprintf("unsupported hook type: %T. Must be a ConfigHookFunc or a StatefulHook", h))
	}

	// Use a unique key for each hook to store it in the instance manager.
	hooks.Set(fmt.Sprintf("%p", hook), hookFunc)
}

// runAfterLoadHooks executes all registered after-load hooks in order.
func runAfterLoadHooks(ctx context.Context, data map[string]any) (map[string]any, error) {
	processedData := data
	var err error

	// All retrieves all registered hooks in a thread-safe manner.
	allHooks := hooks.All()
	for _, hookInstance := range allHooks {
		if hook, ok := hookInstance.(ConfigHookFunc); ok {
			processedData, err = hook(ctx, processedData)
			if err != nil {
				return nil, err
			}
		}
	}

	return processedData, nil
}
