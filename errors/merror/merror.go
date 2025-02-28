package merror

import "github.com/mingzaily/maltose/errors/mcode"

// IEqual 定义了用于比较两个错误是否相等接口
type IEqual interface {
	Error() string
	Equal(target error) bool
}

// ICode 定义了 Code 功能接口
type ICode interface {
	Error() string
	Code() mcode.Code
}

// IStack 定义了 Stack 功能接口
type IStack interface {
	Error() string
	Stack() string
}

// ICause 定义了 Cause 功能接口
type ICause interface {
	Error() string
	Cause() error
}

// ICurrent 定义了 Current 功能接口
type ICurrent interface {
	Error() string
	Current() error
}

// IUnwrap 定义了 Unwrap 功能接口
type IUnwrap interface {
	Error() string
	Unwrap() error
}

const (
	commaSeparatorSpace = ", "
)
