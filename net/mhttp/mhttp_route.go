package mhttp

import (
	"context"
	"fmt"
	"reflect"
	"sort"
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
	s.logger().Infof(ctx, "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)

	if !s.config.PrintRoutes || len(s.routes) == 0 {
		return
	}

	// Prepare data for the table
	type tableRoute struct {
		Method  string
		Path    string
		Handler string
	}
	tableRoutes := make([]tableRoute, 0, len(s.routes))

	// Define column widths
	maxMethod := 6
	maxPath := 4
	maxHandler := 7

	for _, r := range s.routes {
		// skip doc route
		if r.Path == s.config.OpenapiPath || r.Path == s.config.SwaggerPath {
			continue
		}
		// skip health check route
		if s.config.HealthCheck != "" && r.Path == s.config.HealthCheck {
			continue
		}

		// Handler
		handlerType := "Handler"
		if r.Type == routeTypeController {
			controllerName := reflect.TypeOf(r.Controller).Elem().Name()
			handlerType = fmt.Sprintf("%s.%s", controllerName, r.ControllerMethod.Name)
		}
		reqTypeName := "nil"
		if r.ReqType != nil {
			reqTypeName = r.ReqType.String()
		}
		respTypeName := "nil"
		if r.RespType != nil {
			respTypeName = r.RespType.String()
		}
		handlerStr := fmt.Sprintf("%s(%s → %s)", handlerType, reqTypeName, respTypeName)

		// Create table row
		tr := tableRoute{
			Method:  r.Method,
			Path:    r.Path,
			Handler: handlerStr,
		}
		tableRoutes = append(tableRoutes, tr)

		// Update max widths
		if len(tr.Method) > maxMethod {
			maxMethod = len(tr.Method)
		}
		if len(tr.Path) > maxPath {
			maxPath = len(tr.Path)
		}
		if len(tr.Handler) > maxHandler {
			maxHandler = len(tr.Handler)
		}
	}

	// Sort routes by path
	sort.Slice(tableRoutes, func(i, j int) bool {
		return tableRoutes[i].Path < tableRoutes[j].Path
	})

	// Print table
	fmt.Printf("\n┌─ Routes ─%s%s%s\n", strings.Repeat("─", maxMethod), strings.Repeat("─", maxPath), strings.Repeat("─", maxHandler))

	// Header
	headerFormat := fmt.Sprintf("│ %%-%ds │ %%-%ds │ %%-%ds │\n", maxMethod, maxPath, maxHandler)
	fmt.Printf(headerFormat, "METHOD", "PATH", "HANDLER")

	// Separator
	separator := fmt.Sprintf("├─%s─┼─%s─┼─%s─┤\n", strings.Repeat("─", maxMethod), strings.Repeat("─", maxPath), strings.Repeat("─", maxHandler))
	fmt.Print(separator)

	// Body
	for _, tr := range tableRoutes {
		fmt.Printf(headerFormat, tr.Method, tr.Path, tr.Handler)
	}

	// Footer
	footer := fmt.Sprintf("└─%s─┴─%s─┴─%s─┘\n\n", strings.Repeat("─", maxMethod), strings.Repeat("─", maxPath), strings.Repeat("─", maxHandler))
	fmt.Print(footer)
}
