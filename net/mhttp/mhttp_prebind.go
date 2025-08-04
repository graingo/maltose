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
	for _, item := range s.preBindItems {
		var allHandlers []gin.HandlerFunc
		var collectedMiddlewares []MiddlewareFunc

		// Traverse up from the current group to the root, collecting middlewares.
		for g := item.Group; g != nil; g = g.parent {
			collectedMiddlewares = append(collectedMiddlewares, g.middlewares...)
		}

		// Add collected middlewares in correct order to ensure parent middlewares run first.
		for i := 0; i < len(collectedMiddlewares); i++ {
			m := collectedMiddlewares[i]
			allHandlers = append(allHandlers, func(c *gin.Context) {
				m(newRequest(c, s))
			})
		}

		// Add route-level middlewares
		for _, middleware := range item.RouteMiddlewares {
			m := middleware
			allHandlers = append(allHandlers, func(c *gin.Context) {
				m(newRequest(c, s))
			})
		}

		// add final handler function
		finalHandler := func(c *gin.Context) {
			item.HandlerFunc(newRequest(c, s))
		}
		allHandlers = append(allHandlers, finalHandler)

		// register to Gin
		item.Group.ginGroup.Handle(item.Method, item.Path, allHandlers...)
	}

	// clean pre-bind list
	s.preBindItems = nil
}
