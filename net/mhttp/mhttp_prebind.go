package mhttp

import (
	"context"

	"github.com/gin-gonic/gin"
)

// preBindItem 预绑定项
type preBindItem struct {
	Group       *RouterGroup // 路由组
	Method      string       // HTTP 方法
	Path        string       // 路由路径
	HandlerFunc HandlerFunc  // 处理函数
	Type        routeType    // 路由类型
	Controller  any          // 控制器对象
}

// bindRoutes 绑定所有预绑定的路由
func (s *Server) bindRoutes(_ context.Context) {
	for _, item := range s.preBindItems {
		// 转换为 gin handler
		ginHandler := func(c *gin.Context) {
			item.HandlerFunc(newRequest(c, s))
		}

		// 注册到 gin
		if item.Group != nil {
			item.Group.Handle(item.Method, item.Path, ginHandler)
		} else {
			s.Handle(item.Method, item.Path, ginHandler)
		}
	}
}
