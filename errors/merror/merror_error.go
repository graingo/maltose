package merror

import (
	"errors"
	"fmt"

	"github.com/graingo/maltose/errors/mcode"
)

// Error is the error structure.
type Error struct {
	error error
	text  string
	code  mcode.Code
	stack stack
}

// Error implements the Error interface, it returns all error information.
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

// Cause returns the root error.
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

// Current creates and returns the current error.
// If the current error is nil, it returns nil.
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

// Unwrap is an alias function for `Next`.
// It is only for implementing the stdlib errors.Unwrap interface after Go version 1.17.
func (err *Error) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.error
}

// Equal compares two errors for equality.
// Note that in the default error comparison, only when their `code` and `text` are the same, the error is considered the same.
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
