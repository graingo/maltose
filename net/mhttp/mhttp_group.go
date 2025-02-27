package mhttp

import (
	"github.com/gin-gonic/gin"
)

// RouterGroup 路由组
type RouterGroup struct {
	*gin.RouterGroup
	server      *Server
	middlewares []MiddlewareFunc
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
			group.Use(h)
		default:
			s.bindObject(group, h)
		}
	}

	return group
}

// BindHandler 绑定处理函数到路由组
func (g *RouterGroup) BindHandler(method, path string, handler HandlerFunc) {
	// 完整路径
	fullPath := g.BasePath() + path

	// 保存到路由列表
	g.server.routes = append(g.server.routes, Route{
		Method:      method,
		Path:        fullPath,
		HandlerFunc: handler,
		Type:        routeTypeHandler,
	})

	// 添加到预绑定列表
	g.server.preBindItems = append(g.server.preBindItems, preBindItem{
		Group:       g,
		Method:      method,
		Path:        path,
		HandlerFunc: handler,
		Type:        routeTypeHandler,
	})
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
			g.server.bindObject(g, object)
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
		g.middlewares = append(g.middlewares, h)
		g.RouterGroup.Use(handler)
	}
	return g
}
