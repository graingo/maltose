package merror

import (
	"github.com/mingzaily/maltose/errors/mcode"
)

// Code 返回错误码
// 如果错误没有设置错误码，它将返回 Unwrap 方法返回的错误的错误码。
func (err *Error) Code() mcode.Code {
	if err == nil {
		return mcode.CodeNil
	}
	if err.code == mcode.CodeNil {
		return Code(err.Unwrap())
	}
	return err.code
}

// SetCode 设置错误码
func (err *Error) SetCode(code mcode.Code) {
	if err == nil {
		return
	}
	err.code = code
}
