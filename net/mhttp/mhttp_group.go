package mhttp

import (
	"context"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/graingo/maltose/util/mmeta"
)

// RouterGroup is the router group for the server.
type RouterGroup struct {
	server      *Server
	path        string
	ginGroup    *gin.RouterGroup
	middlewares []MiddlewareFunc
}

type RouterGroupOption func(*RouterGroup)

// Group creates a new router group.
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

// Use adds middlewares.
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

// GET registers GET request route.
func (rg *RouterGroup) GET(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("GET", path, handler, middlewares...)
	return rg
}

// POST registers POST request route.
func (rg *RouterGroup) POST(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("POST", path, handler, middlewares...)
	return rg
}

// PUT registers PUT request route.
func (rg *RouterGroup) PUT(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("PUT", path, handler, middlewares...)
	return rg
}

// DELETE registers DELETE request route.
func (rg *RouterGroup) DELETE(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("DELETE", path, handler, middlewares...)
	return rg
}

// HEAD registers HEAD request route.
func (rg *RouterGroup) HEAD(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("HEAD", path, handler, middlewares...)
	return rg
}

// OPTIONS registers OPTIONS request route.
func (rg *RouterGroup) OPTIONS(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("OPTIONS", path, handler, middlewares...)
	return rg
}

// PATCH registers PATCH request route.
func (rg *RouterGroup) PATCH(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares("PATCH", path, handler, middlewares...)
	return rg
}

// Any registers all HTTP methods route.
func (rg *RouterGroup) Any(path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		rg.addRouteWithMiddlewares(method, path, handler, middlewares...)
	}
	return rg
}

// Handle is a general route registration method.
func (rg *RouterGroup) Handle(method, path string, handler HandlerFunc, middlewares ...MiddlewareFunc) *RouterGroup {
	rg.addRouteWithMiddlewares(method, path, handler, middlewares...)
	return rg
}

// BindObject binds the controller object.
func (rg *RouterGroup) BindObject(object any) *RouterGroup {
	return rg.bindObject(object)
}

// addRouteWithMiddlewares is an internal method to add routes with middlewares.
func (rg *RouterGroup) addRouteWithMiddlewares(method, path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	absolutePath := joinPaths(rg.path, path)

	// add to pre-bind list
	rg.server.preBindItems = append(rg.server.preBindItems, preBindItem{
		Group:            rg,
		Method:           method,
		Path:             path,
		HandlerFunc:      handler,
		Type:             routeTypeHandler,
		RouteMiddlewares: middlewares,
	})

	// add to routes list for documentation and printing
	rg.server.routes = append(rg.server.routes, Route{
		Method:      method,
		Path:        absolutePath,
		HandlerFunc: handler,
		Type:        routeTypeHandler,
	})
}

// bindObject handles the route binding of the object (internal method).
func (rg *RouterGroup) bindObject(object any) *RouterGroup {
	typ := reflect.TypeOf(object)
	val := reflect.ValueOf(object)

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)

		// check method signature
		if err := checkMethodSignature(method.Type); err != nil {
			rg.server.Logger().Warnf(context.Background(),
				"method [%s.%s] ignored, %s",
				typ.String(), method.Name, err.Error(),
			)
			continue
		}

		// get request parameter type and metadata
		reqType := method.Type.In(2)
		reqElem := reqType.Elem()
		reqInstance := reflect.New(reqElem).Interface()

		// get route information
		path := mmeta.Get(reqInstance, "path").String()
		httpMethod := mmeta.Get(reqInstance, "method").String()
		if path == "" || httpMethod == "" {
			continue
		}

		// create handler function
		handlerFunc := func(r *Request) {
			req := reflect.New(reqElem).Interface()
			if err := handleRequest(r, method, val, req); err != nil {
				r.Error(err)
			}
		}

		// build full path
		fullPath := joinPaths(rg.path, path)

		// save to routes list
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

		// add to pre-bind list
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

// joinPaths is a helper function to join paths.
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
