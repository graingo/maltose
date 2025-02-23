package mcfg

import "context"

// Adapter 定义配置适配器接口
type Adapter interface {
	// Get 获取指定键的配置值
	Get(ctx context.Context, pattern string) (any, error)

	// Data 获取所有配置数据
	Data(ctx context.Context) (map[string]any, error)

	// Available 检查和返回配置服务是否可用。
	// 可选参数 `resource` 指定某些配置资源。
	Available(ctx context.Context, resource ...string) bool
}
