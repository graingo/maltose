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

	// Step 1: Apply all group-level middlewares exactly once per group using ginGroup.Use().
	for _, item := range s.preBindItems {
		group := item.Group
		if !processedGroups[group] {
			processedGroups[group] = true
			if len(group.middlewares) > 0 {
				var ginMiddlewares []gin.HandlerFunc
				for _, middleware := range group.middlewares {
					m := middleware
					ginMiddlewares = append(ginMiddlewares, func(c *gin.Context) {
						m(newRequest(c, s))
					})
				}
				group.ginGroup.Use(ginMiddlewares...)
			}
		}
	}

	// Step 2: Register all routes with their route-specific middlewares and the final handler.
	// DO NOT add group middlewares here again.
	for _, item := range s.preBindItems {
		var routeHandlers []gin.HandlerFunc

		// ONLY handle route-level middlewares
		for _, middleware := range item.RouteMiddlewares {
			m := middleware
			routeHandlers = append(routeHandlers, func(c *gin.Context) {
				m(newRequest(c, s))
			})
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
