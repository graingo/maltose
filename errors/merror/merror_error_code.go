package merror

import (
	"github.com/graingo/maltose/errors/mcode"
)

// Code returns the error code.
// If the error does not have a set error code, it will return the error code of the error returned by the `Unwrap` method.
func (err *Error) Code() mcode.Code {
	if err == nil {
		return mcode.CodeNil
	}
	if err.code == mcode.CodeNil {
		return Code(err.Unwrap())
	}
	return err.code
}

// SetCode sets the error code.
func (err *Error) SetCode(code mcode.Code) {
	if err == nil {
		return
	}
	err.code = code
}
