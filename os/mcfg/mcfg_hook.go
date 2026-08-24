package mcfg

import (
	"context"
	"fmt"
	"sync"
)

// StatefulHook transforms loaded configuration while retaining state between calls.
type StatefulHook interface {
	Hook(ctx context.Context, data map[string]any) (map[string]any, error)
}

// ConfigHookFunc transforms configuration after an adapter loads it.
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

// RegisterAfterLoadHook registers a process-wide hook that runs after an
// adapter loads configuration. It accepts ConfigHookFunc, its underlying
// function signature, or StatefulHook. Register hooks before the first read.
// Hooks run in registration order.
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
