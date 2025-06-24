package mhttp

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
)

// HandlerFunc defines the basic handler function type.
type HandlerFunc func(*Request)

// handleValidationErrors handles the validation errors.
func handleValidationErrors(r *Request, err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errMsgs []string
		trans := r.GetTranslator()
		for _, e := range validationErrors.Translate(trans) {
			errMsgs = append(errMsgs, e)
		}
		if len(errMsgs) > 0 {
			return merror.NewCode(mcode.CodeValidationFailed, strings.Join(errMsgs, "; "))
		}
	}
	return err
}

// handleRequest handles the request and returns the result.
func handleRequest(r *Request, method reflect.Method, val reflect.Value, req interface{}) error {
	// Parameter binding from URI. We can ignore the error here because
	// not all requests have URI parameters. The main validation for body,
	// query, etc., is handled by ShouldBind below.
	_ = r.ShouldBindUri(req)

	// parameter binding from query, form, body, etc.
	if err := r.ShouldBind(req); err != nil {
		return handleValidationErrors(r, err)
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
	if !typ.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return merror.New("first parameter should be context.Context")
	}

	// check if the third parameter is request parameter
	reqType := typ.In(2)
	if reqType.Kind() != reflect.Ptr {
		return merror.New("request parameter should be pointer type")
	}
	if !strings.HasSuffix(reqType.Elem().Name(), "Req") {
		return merror.New("request parameter should end with 'Req'")
	}

	// check if the first return value is response parameter
	resType := typ.Out(0)
	if resType.Kind() != reflect.Ptr {
		return merror.New("response parameter should be pointer type")
	}
	if !strings.HasSuffix(resType.Elem().Name(), "Res") {
		return merror.New("response parameter should end with 'Res'")
	}

	// check if the second return value is error
	if !typ.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return merror.New("second return value should be error")
	}

	return nil
}
