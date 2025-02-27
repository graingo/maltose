package mhttp

import "reflect"

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
