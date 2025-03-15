package mhttp

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type routeType int

const (
	routeTypeHandler routeType = iota
	routeTypeController
)

// Route is the route information.
type Route struct {
	Method           string
	Path             string
	HandlerFunc      HandlerFunc
	Type             routeType
	Controller       any            // controller object
	ControllerMethod reflect.Method // controller method
	ReqType          reflect.Type   // request parameter type
	RespType         reflect.Type   // response type
}

func (s *Server) Routes() []Route {
	return s.routes
}

func (s *Server) printRoute(ctx context.Context) {
	// print server info
	s.Logger().Infof(ctx, "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)

	// print route info
	fmt.Printf("\n%s\n", strings.Repeat("-", 100))
	fmt.Printf("%-40s | %-7s | %-20s | %-30s\n",
		"PATH", "METHOD", "HANDLER TYPE", "REQUEST/RESPONSE")
	fmt.Printf("%s\n", strings.Repeat("-", 100))

	for _, route := range s.routes {
		// skip doc route
		if route.Path == s.config.OpenapiPath || route.Path == s.config.SwaggerPath {
			continue
		}

		// determine route type description
		handlerType := "Handler"
		if route.Type == routeTypeController {
			controllerName := reflect.TypeOf(route.Controller).Elem().Name()
			handlerType = fmt.Sprintf("%s.%s", controllerName, route.ControllerMethod.Name)
		}

		// get request and response type names
		reqTypeName := "nil"
		respTypeName := "nil"

		if route.ReqType != nil {
			if route.ReqType.Kind() == reflect.Ptr {
				reqTypeName = "*" + route.ReqType.Elem().Name()
			} else {
				reqTypeName = route.ReqType.Name()
			}
		}

		if route.RespType != nil {
			if route.RespType.Kind() == reflect.Ptr {
				respTypeName = "*" + route.RespType.Elem().Name()
			} else {
				respTypeName = route.RespType.Name()
			}
		}

		// print route info
		fmt.Printf("%-40s | %-7s | %-20s | %s → %s\n",
			route.Path,
			route.Method,
			handlerType,
			reqTypeName,
			respTypeName,
		)
	}
	fmt.Printf("%s\n\n", strings.Repeat("-", 100))
}
