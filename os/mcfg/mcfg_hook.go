package mcfg

import (
	"context"
	"fmt"
	"sync"
)

// StatefulHook is an interface for hooks that need to maintain state across multiple calls.
// This is useful for caching results of expensive operations, like fetching remote configuration.
type StatefulHook interface {
	Hook(ctx context.Context, data map[string]any) (map[string]any, error)
}

type ConfigHookFunc func(ctx context.Context, data map[string]any) (map[string]any, error)

type hookRegistry struct {
	mu      sync.RWMutex
	keys    map[string]struct{}
	ordered []ConfigHookFunc
}

var hooks = &hookRegistry{keys: make(map[string]struct{})}

func (r *hookRegistry) register(key string, hook ConfigHookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.keys[key]; exists {
		return
	}
	r.keys[key] = struct{}{}
	r.ordered = append(r.ordered, hook)
}

func (r *hookRegistry) all() []ConfigHookFunc {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]ConfigHookFunc(nil), r.ordered...)
}

func (r *hookRegistry) count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.ordered)
}

func (r *hookRegistry) clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.keys = make(map[string]struct{})
	r.ordered = nil
}

// RegisterAfterLoadHook registers a hook to be executed after configuration is loaded.
// It accepts either a function with the signature `func(context.Context, map[string]any) (map[string]any, error)`
// or an implementation of the `StatefulHook` interface.
// Using a `StatefulHook` is the recommended way to handle expensive operations that should only run once (e.g., fetching remote config),
// as it allows caching within the hook's state.
// Each hook is stored with a unique key to prevent duplicate registrations.
func RegisterAfterLoadHook(hook any) {
	var hookFunc ConfigHookFunc
	switch h := hook.(type) {
	case ConfigHookFunc:
		hookFunc = h
	case func(context.Context, map[string]any) (map[string]any, error):
		hookFunc = h
	case StatefulHook:
		hookFunc = h.Hook
	default:
		panic(fmt.Sprintf("unsupported hook type: %T. Must be a ConfigHookFunc or a StatefulHook", h))
	}

	hooks.register(fmt.Sprintf("%p", hook), hookFunc)
}

// ClearHooks removes all registered hooks.
// This is intended for testing purposes only.
func ClearHooks() {
	hooks.clear()
}

// runAfterLoadHooks executes all registered after-load hooks in order.
func runAfterLoadHooks(ctx context.Context, data map[string]any) (map[string]any, error) {
	processedData := data
	var err error

	for _, hook := range hooks.all() {
		processedData, err = hook(ctx, processedData)
		if err != nil {
			return nil, err
		}
	}

	return processedData, nil
}
