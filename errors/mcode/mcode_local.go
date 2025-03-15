package mcode

import "fmt"

// localCode is the implementation of the error code.
type localCode struct {
	code    int
	message string
	detail  any
}

// Code returns the error code.
func (c localCode) Code() int {
	return c.code
}

// Message returns the error message.
func (c localCode) Message() string {
	return c.message
}

// Detail returns the detailed information of the current error code, mainly for the extended fields of the error code.
func (c localCode) Detail() any {
	return c.detail
}

// String returns the string representation of the error code.
func (c localCode) String() string {
	if c.detail != nil {
		return fmt.Sprintf(`%d:%s %v`, c.code, c.message, c.detail)
	}
	if c.message != "" {
		return fmt.Sprintf(`%d:%s`, c.code, c.message)
	}
	return fmt.Sprintf(`%d`, c.code)
}
