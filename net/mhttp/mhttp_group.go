package mhttp

import (
	"context"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/graingo/maltose/util/mmeta"
)

// RouterGroup 路由组
type RouterGroup struct {
	server      *Server
	path        string
	ginGroup    *gin.RouterGroup
	middlewares []MiddlewareFunc
}

type RouterGroupOption func(*RouterGroup)

// Group 创建新的路由组
func (rg *RouterGroup) Group(path string, handlers ...RouterGroupOption) *RouterGroup {
	group := &RouterGroup{
		server:   rg.server,
		path:     joinPaths(rg.path, path),
		ginGroup: rg.ginGroup.Group(path),
	}

	for _, handler := range handlers {
		handler(group)
	}

	return group
}

// Use 添加中间件
func (rg *RouterGroup) Use(middlewares []MiddlewareFunc, handlers ...RouterGroupOption) *RouterGroup {
	if rg.middlewares == nil {
		rg.middlewares = make([]MiddlewareFunc, 0, len(middlewares))
	}
	rg.middlewares = append(rg.middlewares, middlewares...)

	for _, handler := range handlers {
		handler(rg)
	}

	return rg
}

// GET 注册 GET 请求路由
func (rg *RouterGroup) GET(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("GET", path, handler)
	return rg
}

// POST 注册 POST 请求路由
func (rg *RouterGroup) POST(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("POST", path, handler, middlewares...)
	return rg
}

// PUT 注册 PUT 请求路由
func (rg *RouterGroup) PUT(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("PUT", path, handler, middlewares...)
	return rg
}

// DELETE 注册 DELETE 请求路由
func (rg *RouterGroup) DELETE(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("DELETE", path, handler, middlewares...)
	return rg
}

// HEAD 注册 HEAD 请求路由
func (rg *RouterGroup) HEAD(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("HEAD", path, handler, middlewares...)
	return rg
}

// OPTIONS 注册 OPTIONS 请求路由
func (rg *RouterGroup) OPTIONS(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("OPTIONS", path, handler, middlewares...)
	return rg
}

// PATCH 注册 PATCH 请求路由
func (rg *RouterGroup) PATCH(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("PATCH", path, handler, middlewares...)
	return rg
}

// Any 注册所有 HTTP 方法路由
func (rg *RouterGroup) Any(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		rg.addRouteWithMiddlewares(method, path, handler, middlewares...)
	}
	return rg
}

// Handle 通用路由注册方法
func (rg *RouterGroup) Handle(method, path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares(method, path, handler, middlewares...)
	return rg
}

// BindObject 绑定控制器对象
func (rg *RouterGroup) BindObject(object any) *RouterGroup {
	return rg.bindObject(object)
}

// 内部方法，添加路由
func (rg *RouterGroup) addRouteWithMiddlewares(method, path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	absolutePath := joinPaths(rg.path, path)

	// 添加到预绑定列表
	rg.server.preBindItems = append(rg.server.preBindItems, preBindItem{
		Group:            rg,
		Method:           method,
		Path:             path,
		HandlerFunc:      handler,
		Type:             routeTypeHandler,
		RouteMiddlewares: middlewares,
	})

	// 添加到路由列表用于文档和打印
	rg.server.routes = append(rg.server.routes, Route{
		Method:      method,
		Path:        absolutePath,
		HandlerFunc: handler,
		Type:        routeTypeHandler,
	})
}

// bindObject 处理对象的路由绑定 (内部方法)
func (rg *RouterGroup) bindObject(object any) *RouterGroup {
	typ := reflect.TypeOf(object)
	val := reflect.ValueOf(object)

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)

		// 检查方法签名
		if err := checkMethodSignature(method.Type); err != nil {
			rg.server.Logger().Warnf(context.Background(),
				"method [%s.%s] ignored, %s",
				typ.String(), method.Name, err.Error(),
			)
			continue
		}

		// 获取请求参数类型和元数据
		reqType := method.Type.In(2)
		reqElem := reqType.Elem()
		reqInstance := reflect.New(reqElem).Interface()

		// 获取路由信息
		path := mmeta.Get(reqInstance, "path").String()
		httpMethod := mmeta.Get(reqInstance, "method").String()
		if path == "" || httpMethod == "" {
			continue
		}

		// 创建处理函数
		handlerFunc := func(r *Request) {
			req := reflect.New(reqElem).Interface()
			if err := handleRequest(r, method, val, req); err != nil {
				r.Error(err)
			}
		}

		// 构建完整路径
		fullPath := joinPaths(rg.path, path)

		// 保存到路由列表
		rg.server.routes = append(rg.server.routes, Route{
			Method:           httpMethod,
			Path:             fullPath,
			HandlerFunc:      handlerFunc,
			Type:             routeTypeController,
			Controller:       object,
			ControllerMethod: method,
			ReqType:          reqType,
			RespType:         method.Type.Out(0),
		})

		// 添加到预绑定列表
		rg.server.preBindItems = append(rg.server.preBindItems, preBindItem{
			Group:       rg,
			Method:      httpMethod,
			Path:        path,
			HandlerFunc: handlerFunc,
			Type:        routeTypeController,
			Controller:  object,
		})
	}

	return rg
}

// 辅助函数，连接路径
func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := absolutePath
	if absolutePath != "/" {
		finalPath += "/"
	}

	if relativePath[0] == '/' {
		finalPath += relativePath[1:]
	} else {
		finalPath += relativePath
	}

	return finalPath
}
