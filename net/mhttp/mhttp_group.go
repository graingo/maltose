package mhttp

import (
	"github.com/gin-gonic/gin"
)

// GroupHandler 定义路由组处理函数类型
type GroupHandler func(group *RouterGroup)

// RouterGroup 路由组
type RouterGroup struct {
	*gin.RouterGroup
	server     *Server
	middleware []MiddlewareFunc
}

// Group 创建路由组
func (s *Server) Group(prefix string, handlers ...any) *RouterGroup {
	group := &RouterGroup{
		RouterGroup: s.Engine.Group(prefix),
		server:      s,
	}

	// 处理中间件
	for _, handler := range handlers {
		switch h := handler.(type) {
		case MiddlewareFunc:
			group.RouterGroup.Use(func(c *gin.Context) {
				r := newRequest(c, s)
				h(r)
			})
		case GroupHandler:
			h(group)
		}
	}

	return group
}

// BindHandler 绑定处理函数到路由组
func (g *RouterGroup) BindHandler(method, path string, handler HandlerFunc) *RouterGroup {
	g.Handle(method, path, func(c *gin.Context) {
		handler(RequestFromCtx(c))
	})
	return g
}

// Bind 绑定控制器到路由组
func (g *RouterGroup) Bind(objects ...any) *RouterGroup {
	for _, object := range objects {
		switch h := object.(type) {
		case MiddlewareFunc:
			// 处理中间件
			g.Use(h)
		default:
			// 处理控制器对象
			g.server.bindObject(g.RouterGroup, object)
		}
	}
	return g
}

// Middleware 为路由组添加中间件
func (g *RouterGroup) Use(handlers ...MiddlewareFunc) *RouterGroup {
	for _, h := range handlers {
		handler := func(c *gin.Context) {
			r := newRequest(c, g.server)
			h(r)
		}
		g.middleware = append(g.middleware, h)
		g.RouterGroup.Use(handler)
	}
	return g
}
