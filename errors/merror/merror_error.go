package merror

import (
	"errors"
	"fmt"

	"github.com/mingzaily/maltose/errors/mcode"
)

// Error 错误结构体
type Error struct {
	error error
	text  string
	code  mcode.Code
	stack stack
}

// Error 实现了 Error 接口，它返回所有的错误信息。
func (err *Error) Error() string {
	if err == nil {
		return ""
	}
	errStr := err.text
	if errStr == "" && err.code != nil {
		errStr = err.code.Message()
	}
	if err.error != nil {
		if errStr != "" {
			errStr += ": "
		}
		errStr += err.error.Error()
	}
	return errStr
}

// Cause 返回根错误
func (err *Error) Cause() error {
	if err == nil {
		return nil
	}
	loop := err
	for loop != nil {
		if loop.error != nil {
			if e, ok := loop.error.(*Error); ok {
				// Internal Error struct.
				loop = e
			} else if e, ok := loop.error.(ICause); ok {
				// Other Error that implements ApiCause interface.
				return e.Cause()
			} else {
				return loop.error
			}
		} else {
			// return loop
			//
			// To be compatible with Case of https://github.com/pkg/errors.
			return errors.New(loop.text)
		}
	}
	return nil
}

// Current 创建并返回当前错误。
// 如果当前错误为 nil，则返回 nil。
func (err *Error) Current() error {
	if err == nil {
		return nil
	}
	return &Error{
		error: nil,
		stack: err.stack,
		text:  err.text,
		code:  err.code,
	}
}

// Unwrap 别名函数 `Next`。
// 它只是为了实现 Go 版本 1.17 之后的 stdlib errors.Unwrap 接口。
func (err *Error) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.error
}

// Equal 比较两个错误是否相同。
// 请注意，在默认的错误比较中，只有当它们的 `code` 和 `text` 相同时，错误才被认为是相同的。
func (err *Error) Equal(target error) bool {
	if err == target {
		return true
	}
	// Code should be the same.
	// Note that if both errors have `nil` code, they are also considered equal.
	if err.code != Code(target) {
		return false
	}
	// Text should be the same.
	if err.text != fmt.Sprintf(`%-s`, target) {
		return false
	}
	return true
}
