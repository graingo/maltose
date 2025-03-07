package mhttp

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
)

// HandlerFunc 定义基础处理函数类型
type HandlerFunc func(*Request)

// handleRequest 处理请求并返回结果
func handleRequest(r *Request, method reflect.Method, val reflect.Value, req interface{}) error {
	// 参数绑定
	if err := r.ShouldBind(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errMsgs []string
			for _, e := range validationErrors.Translate(r.GetTranslator()) {
				errMsgs = append(errMsgs, e)
			}
			if len(errMsgs) > 0 {
				return merror.NewCode(mcode.CodeValidationFailed, errMsgs[0])
			}
		}
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
