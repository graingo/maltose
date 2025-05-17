package mcfg

import (
	"context"
	"sync"
)

type ConfigHookFunc func(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)

var (
	// 配置加载后处理钩子
	afterLoadHooks     []ConfigHookFunc
	afterLoadHookMutex sync.RWMutex
)

// RegisterAfterLoadHook 注册一个配置加载后的处理钩子
func RegisterAfterLoadHook(hook ConfigHookFunc) {
	afterLoadHookMutex.Lock()
	defer afterLoadHookMutex.Unlock()

	afterLoadHooks = append(afterLoadHooks, hook)
}

// runAfterLoadHooks 执行所有配置加载后的钩子
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
