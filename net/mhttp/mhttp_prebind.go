package mhttp

import (
	"context"

	"github.com/gin-gonic/gin"
)

// preBindItem 预绑定项
type preBindItem struct {
	Group            *RouterGroup
	Method           string
	Path             string
	HandlerFunc      HandlerFunc
	Type             routeType
	Controller       interface{}
	RouteMiddlewares []MiddlewareFunc
}

// bindRoutes 绑定所有预绑定的路由
func (s *Server) bindRoutes(_ context.Context) {
	processedGroups := make(map[*RouterGroup]bool)

	for _, item := range s.preBindItems {
		group := item.Group
		if !processedGroups[group] {
			processedGroups[group] = true
			for _, middleware := range group.middlewares {
				ginMiddleware := func(c *gin.Context) {
					middleware(newRequest(c, s))
				}
				group.ginGroup.Use(ginMiddleware)
			}
		}
	}

	// 第二步：注册所有路由及其路由级中间件
	for _, item := range s.preBindItems {
		var routeHandlers []gin.HandlerFunc

		// 只处理路由级中间件
		for _, middleware := range item.RouteMiddlewares {
			ginMiddleware := func(c *gin.Context) {
				middleware(newRequest(c, s))
			}
			routeHandlers = append(routeHandlers, ginMiddleware)
		}

		// 添加最终处理函数
		finalHandler := func(c *gin.Context) {
			item.HandlerFunc(newRequest(c, s))
		}
		routeHandlers = append(routeHandlers, finalHandler)

		// 注册到 Gin
		item.Group.ginGroup.Handle(item.Method, item.Path, routeHandlers...)
	}

	// 清理预绑定列表
	s.preBindItems = nil

	// 清理中间件引用以帮助垃圾回收
	for group := range processedGroups {
		group.middlewares = nil
	}
}
