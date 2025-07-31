package merror

import (
	"fmt"

	"github.com/graingo/maltose/errors/mcode"
)

// New creates a new error.
// Example: err := merror.New("username cannot be empty")
func New(text string) error {
	return &Error{
		stack: callers(),
		text:  text,
		code:  mcode.CodeNil,
	}
}

// Newf creates a new error.
// Example: err := merror.Newf("username %s cannot be empty", admin)
func Newf(format string, a ...any) error {
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, a...),
		code:  mcode.CodeNil,
	}
}

// Wrap wraps an error.
// Example: err := merror.Wrap(err, "username cannot be empty")
func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{
		stack: callers(),
		text:  text,
		error: err,
		code:  mcode.CodeNil,
	}
}

// Wrapf wraps an error.
// Example: err := merror.Wrapf(err, "username %s cannot be empty", admin)
func Wrapf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	return &Error{
		stack: callers(),
		text:  fmt.Sprintf(format, a...),
		error: err,
		code:  mcode.CodeNil,
	}
}
