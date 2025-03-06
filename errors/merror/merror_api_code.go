package merror

import (
	"fmt"
	"strings"

	"github.com/savorelle/maltose/errors/mcode"
)

// NewCode 创建一个带指定错误码的新错误
// 示例：err := merror.NewCode(mcode.ValidationError)
func NewCode(code mcode.Code, text ...string) error {
	return &Error{
		stack: callers(),
		text:  strings.Join(text, commaSeparatorSpace),
		code:  code,
	}
}

// NewCodef 创建一个带指定错误码的新错误
// 示例：err := merror.NewCodef(mcode.ValidationError, "用户名%s不能为空", admin)
func NewCodef(code mcode.Code, format string, args ...any) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  code,
	}
}

// WrapCode 用于包装一个错误，并附加指定的错误码
// 示例：err := merror.WrapCode(err, mcode.ValidationError)
func WrapCode(err error, code mcode.Code, text ...string) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  strings.Join(text, commaSeparatorSpace),
		code:  code,
	}
}

// WrapCodef 用于包装一个错误，并附加指定的错误码和格式化文本
// 示例：err := merror.WrapCodef(err, mcode.ValidationError, "用户名%s不能为空", admin)
func WrapCodef(err error, code mcode.Code, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &Error{
		error: err,
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  code,
	}
}

// Code 用于获取错误的错误码
// 如果没有错误代码，也没有实现接口代码，则返回 CodeNil
func Code(err error) mcode.Code {
	if err == nil {
		return mcode.CodeNil
	}
	if e, ok := err.(ICode); ok {
		return e.Code()
	}
	if e, ok := err.(IUnwrap); ok {
		return Code(e.Unwrap())
	}
	return mcode.CodeNil
}
