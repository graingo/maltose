package mhttp

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/graingo/maltose/net/mhttp/contract"

	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
)

// HandlerFunc defines the basic handler function type.
type HandlerFunc func(*Request)

// handleRequest handles the request and returns the result.
func handleRequest(r *Request, method reflect.Method, val reflect.Value, req interface{}) error {
	paths := make(map[string]string, len(r.Params))
	for _, p := range r.Params {
		paths[p.Key] = p.Value
	}
	input, err := r.operation.Bind(r.Request, paths, req)
	r.input = input
	if err != nil {
		var validation *contract.ValidationError
		if errors.As(err, &validation) {
			return merror.WrapCode(err, mcode.CodeValidationFailed, "request validation failed")
		}
		return err
	}

	// call method
	results := method.Func.Call([]reflect.Value{
		val,
		reflect.ValueOf(r.Request.Context()),
		reflect.ValueOf(req),
	})

	// handle return value
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	if results[0].IsNil() && r.SuccessStatus() != 204 {
		return merror.New("controller returned a nil success response")
	}

	// set response to Request for middleware usage
	response := results[0].Interface()
	r.SetHandlerResponse(response)

	return nil
}

// checkMethodSignature checks the method signature.
func checkMethodSignature(typ reflect.Type) error {
	// check parameter number and return value number
	if typ.NumIn() != 3 || typ.NumOut() != 2 {
		return merror.New("invalid method signature, required: func(*Controller) (context.Context, *XxxReq) (*XxxRes, error)")
	}

	// check if the second parameter is context.Context
	if typ.In(1) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		return merror.New("first parameter should be context.Context")
	}

	// check if the third parameter is request parameter
	reqType := typ.In(2)
	if reqType.Kind() != reflect.Pointer {
		return merror.New("request parameter should be pointer type")
	}
	if !strings.HasSuffix(reqType.Elem().Name(), "Req") {
		return merror.New("request parameter should end with 'Req'")
	}

	// check if the first return value is response parameter
	resType := typ.Out(0)
	if resType.Kind() != reflect.Pointer {
		return merror.New("response parameter should be pointer type")
	}
	if !strings.HasSuffix(resType.Elem().Name(), "Res") {
		return merror.New("response parameter should end with 'Res'")
	}

	// check if the second return value is error
	if typ.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return merror.New("second return value should be error")
	}

	return nil
}
