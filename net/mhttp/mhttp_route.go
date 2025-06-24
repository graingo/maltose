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
	s.Logger().Infof(ctx, "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)

	if !s.config.PrintRoutes || len(s.routes) == 0 {
		return
	}

	// Group routes by prefix and subpath
	// groups[prefix][subpath] = []Route
	groups := make(map[string]map[string][]Route)
	for _, route := range s.routes {
		// skip doc route
		if route.Path == s.config.OpenapiPath || route.Path == s.config.SwaggerPath {
			continue
		}

		parts := strings.Split(strings.Trim(route.Path, "/"), "/")
		prefix := "/" + parts[0]
		subpath := "/"

		if len(parts) > 1 {
			subpath += strings.Join(parts[1:], "/")
		}

		if _, ok := groups[prefix]; !ok {
			groups[prefix] = make(map[string][]Route)
		}
		groups[prefix][subpath] = append(groups[prefix][subpath], route)
	}

	// Sort and print routes
	sortedPrefixes := make([]string, 0, len(groups))
	for p := range groups {
		sortedPrefixes = append(sortedPrefixes, p)
	}
	sort.Strings(sortedPrefixes)

	// Header
	fmt.Printf("\n ┌── Routes ──────────────────────────────────────────────────────────────────────\n")
	fmt.Printf(" │\n")

	for _, prefix := range sortedPrefixes {
		fmt.Printf(" │  ▼ %s\n", prefix)
		subRoutes := groups[prefix]
		sortedSubPaths := make([]string, 0, len(subRoutes))
		for sp := range subRoutes {
			sortedSubPaths = append(sortedSubPaths, sp)
		}
		sort.Strings(sortedSubPaths)

		for i, subpath := range sortedSubPaths {
			isLastSubPath := (i == len(sortedSubPaths)-1)
			subPathSymbol := "├"
			routeLinePrefix := "│    │"
			if isLastSubPath {
				subPathSymbol = "└"
				routeLinePrefix = "│     "
			}
			fmt.Printf(" │    %s── %s\n", subPathSymbol, subpath)

			routes := subRoutes[subpath]
			sort.Slice(routes, func(i, j int) bool {
				return routes[i].Method < routes[j].Method
			})

			for j, route := range routes {
				isLastRoute := (j == len(routes)-1)
				routeSymbol := "├"
				if isLastRoute {
					routeSymbol = "└"
				}

				handlerType := "Handler"
				if route.Type == routeTypeController {
					controllerName := reflect.TypeOf(route.Controller).Elem().Name()
					handlerType = fmt.Sprintf("%s.%s", controllerName, route.ControllerMethod.Name)
				}
				reqTypeName := "nil"
				if route.ReqType != nil {
					reqTypeName = route.ReqType.String()
				}
				respTypeName := "nil"
				if route.RespType != nil {
					respTypeName = route.RespType.String()
				}

				handlerStr := fmt.Sprintf("%s(%s → %s)", handlerType, reqTypeName, respTypeName)
				fmt.Printf(" %s %s─ %-7s → %s\n", routeLinePrefix, routeSymbol, route.Method, handlerStr)
			}
		}
	}

	// Footer
	fmt.Printf(" │\n")
	fmt.Printf(" └──────────────────────────────────────────────────────────────────────────────────\n\n")
}
