package merror

import (
	"fmt"

	"github.com/graingo/maltose/errors/mcode"
)

// New 创建一个新的错误
// 例子：err := merror.New("用户名不能为空")
func New(text string) error {
	return &Error{
		stack: callers(),
		text:  text,
		code:  mcode.CodeNil,
	}
}

// Newf 创建一个新的错误
// 例子：err := merror.Newf("用户名%s不能为空", admin)
func Newf(format string, a ...any) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, a...),
		code:  mcode.CodeNil,
	}
}

// Wrap 包装一个错误
// 例子：err := merror.Wrap(err, "用户名不能为空")
func Wrap(err error, text string) error {
	return &Error{
		stack: callers(),
		text:  text,
		error: err,
		code:  mcode.CodeNil,
	}
}

// Wrapf 包装一个错误
// 例子：err := merror.Wrapf(err, "用户名%s不能为空", admin)
func Wrapf(err error, format string, a ...any) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, a...),
		error: err,
		code:  mcode.CodeNil,
	}
}
