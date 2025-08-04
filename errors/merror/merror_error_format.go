package merror

import (
	"fmt"
	"io"
)

// Format implements the fmt.Formatter interface, it can format the error information.
//
// Format specifiers:
//
//	%s: error information
//	+v: error information and stack information
func (err *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		switch {
		case s.Flag('-'):
			_, _ = io.WriteString(s, err.Error())
		case s.Flag('+'):
			if verb == 's' {
				_, _ = io.WriteString(s, err.Stack())
			} else {
				_, _ = io.WriteString(s, err.Error()+"\n"+err.Stack())
			}
		default:
			_, _ = io.WriteString(s, err.Error())
		}
	}
}
