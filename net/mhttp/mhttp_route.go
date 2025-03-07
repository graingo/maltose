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

type Route struct {
	Method           string
	Path             string
	HandlerFunc      HandlerFunc
	Type             routeType
	Controller       any            // 控制器对象
	ControllerMethod reflect.Method // 控制器方法
	ReqType          reflect.Type   // 请求参数类型
	RespType         reflect.Type   // 响应类型
}

func (s *Server) Routes() []Route {
	return s.routes
}

func (s *Server) printRoute(ctx context.Context) {
	// 打印服务信息
	s.Logger().Infof(ctx, "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)

	// 打印路由信息
	fmt.Printf("\n%s\n", strings.Repeat("-", 100))
	fmt.Printf("%-40s | %-7s | %-20s | %-30s\n",
		"PATH", "METHOD", "HANDLER TYPE", "REQUEST/RESPONSE")
	fmt.Printf("%s\n", strings.Repeat("-", 100))

	for _, route := range s.routes {
		// 跳过文档路由
		if route.Path == s.config.OpenapiPath || route.Path == s.config.SwaggerPath {
			continue
		}

		// 确定路由类型描述
		handlerType := "Handler"
		if route.Type == routeTypeController {
			controllerName := reflect.TypeOf(route.Controller).Elem().Name()
			handlerType = fmt.Sprintf("%s.%s", controllerName, route.ControllerMethod.Name)
		}

		// 获取请求和响应类型名称
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

		// 打印路由信息
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
