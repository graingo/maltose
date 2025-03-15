package merror

import (
	"fmt"
	"strings"

	"github.com/graingo/maltose/errors/mcode"
)

// NewCode creates a new error with the specified error code.
// Example: err := merror.NewCode(mcode.ValidationError)
func NewCode(code mcode.Code, text ...string) error {
	return &Error{
		stack: callers(),
		text:  strings.Join(text, commaSeparatorSpace),
		code:  code,
	}
}

// NewCodef creates a new error with the specified error code.
// Example: err := merror.NewCodef(mcode.ValidationError, "username %s cannot be empty", admin)
func NewCodef(code mcode.Code, format string, args ...any) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, args...),
		code:  code,
	}
}

// WrapCode wraps an error and appends the specified error code.
// Example: err := merror.WrapCode(err, mcode.ValidationError)
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

// WrapCodef wraps an error and appends the specified error code and formatted text.
// Example: err := merror.WrapCodef(err, mcode.ValidationError, "username %s cannot be empty", admin)
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

// Code gets the error code of the error.
// If there is no error code, and the error does not implement the ICode interface, it returns CodeNil.
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
