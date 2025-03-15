package mhttp

import (
	"context"

	"github.com/gin-gonic/gin"
)

// preBindItem is the pre-binding item.
type preBindItem struct {
	Group            *RouterGroup
	Method           string
	Path             string
	HandlerFunc      HandlerFunc
	Type             routeType
	Controller       interface{}
	RouteMiddlewares []MiddlewareFunc
}

// bindRoutes binds all pre-bound routes.
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

	// second step: register all routes and their route-level middlewares
	for _, item := range s.preBindItems {
		var routeHandlers []gin.HandlerFunc

		// only handle route-level middlewares
		for _, middleware := range item.RouteMiddlewares {
			ginMiddleware := func(c *gin.Context) {
				middleware(newRequest(c, s))
			}
			routeHandlers = append(routeHandlers, ginMiddleware)
		}

		// add final handler function
		finalHandler := func(c *gin.Context) {
			item.HandlerFunc(newRequest(c, s))
		}
		routeHandlers = append(routeHandlers, finalHandler)

		// register to Gin
		item.Group.ginGroup.Handle(item.Method, item.Path, routeHandlers...)
	}

	// clean pre-bind list
	s.preBindItems = nil

	// clean middleware references to help garbage collection
	for group := range processedGroups {
		group.middlewares = nil
	}
}
