package mhttp

import (
	"context"
	"testing"

	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/assert"
)

// 测试控制器
type OptimizedTestController struct{}

type OptimizedTestReq struct {
	mmeta.Meta `path:"/optimized" method:"post"`
	Name       string `json:"name"`
	Count      int    `json:"count"`
}

type OptimizedTestRes struct {
	Message string `json:"message"`
	Total   int    `json:"total"`
}

func (c *OptimizedTestController) OptimizedTest(_ context.Context, req *OptimizedTestReq) (*OptimizedTestRes, error) {
	return &OptimizedTestRes{
		Message: "Hello " + req.Name,
		Total:   req.Count * 2,
	}, nil
}

// 基准测试：对比优化前后的性能
func BenchmarkHandler(b *testing.B) {
	server := New()
	controller := &OptimizedTestController{}

	// 绑定控制器
	server.Bind(controller)

	// 创建测试请求
	req := &OptimizedTestReq{
		Name:  "test",
		Count: 10000,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 模拟请求处理
			resp, err := controller.OptimizedTest(context.Background(), req)
			assert.NoError(b, err)
			assert.NotNil(b, resp)
		}
	})
}
