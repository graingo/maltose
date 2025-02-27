package mhttp

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/mingzaily/maltose/util/mmeta"
)

func (s *Server) doPrintRoute(ctx context.Context) {
	// 打印服务信息
	s.Logger().Infof(ctx, "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)
	// 打印路由信息
	fmt.Printf("%s\n", strings.Repeat("-", 60))
	fmt.Printf("%-10s | %-7s | %-15s \n", "ADDRESS", "METHOD", "ROUTE")
	routes := s.Engine.Routes()
	for _, route := range routes {
		if route.Path == s.config.OpenapiPath || route.Path == s.config.SwaggerPath {
			continue
		}
		fmt.Printf("%-10s | %-7s | %-15s \n",
			s.config.Address,
			route.Method,
			route.Path,
		)
	}
	fmt.Printf("%s\n\n", strings.Repeat("-", 60))
}

// bindObject 处理对象的路由绑定
func (s *Server) bindObject(group *RouterGroup, object any) {
	typ := reflect.TypeOf(object)
	val := reflect.ValueOf(object)

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)

		// 检查方法签名
		if err := checkMethodSignature(method.Type); err != nil {
			s.Logger().Warnf(context.Background(),
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
				panic(err)
			}
		}

		// 构建完整路径
		fullPath := path
		if group != nil {
			fullPath = group.BasePath() + path
		}

		// 保存到路由列表
		s.routes = append(s.routes, Route{
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
		s.preBindItems = append(s.preBindItems, preBindItem{
			Group:       group,
			Method:      httpMethod,
			Path:        path,
			HandlerFunc: handlerFunc,
			Type:        routeTypeController,
			Controller:  object,
		})
	}
}

// handleRequest 处理请求并返回结果
func handleRequest(r *Request, method reflect.Method, val reflect.Value, req interface{}) error {
	// 参数绑定
	if err := r.ShouldBind(req); err != nil {
		return err
	}

	// 调用方法
	results := method.Func.Call([]reflect.Value{
		val,
		reflect.ValueOf(r.Request.Context()),
		reflect.ValueOf(req),
	})

	// 处理返回值
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	// 设置响应到 Request 中供中间件使用
	response := results[0].Interface()
	r.SetHandlerResponse(response)

	return nil
}

// checkMethodSignature 检查方法签名是否符合要求
func checkMethodSignature(typ reflect.Type) error {
	// 检查参数数量和返回值数量
	if typ.NumIn() != 3 || typ.NumOut() != 2 {
		return fmt.Errorf("invalid method signature, required: func(*Controller) (context.Context, *XxxReq) (*XxxRes, error)")
	}

	// 检查第二个参数是否为 context.Context
	if !typ.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return fmt.Errorf("first parameter should be context.Context")
	}

	// 检查第三个参数（请求参数）
	reqType := typ.In(2)
	if reqType.Kind() != reflect.Ptr {
		return fmt.Errorf("request parameter should be pointer type")
	}
	if !strings.HasSuffix(reqType.Elem().Name(), "Req") {
		return fmt.Errorf("request parameter should end with 'Req'")
	}

	// 检查第一个返回值（响应参数）
	resType := typ.Out(0)
	if resType.Kind() != reflect.Ptr {
		return fmt.Errorf("response parameter should be pointer type")
	}
	if !strings.HasSuffix(resType.Elem().Name(), "Res") {
		return fmt.Errorf("response parameter should end with 'Res'")
	}

	// 检查第二个返回值是否为 error
	if !typ.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("second return value should be error")
	}

	return nil
}
