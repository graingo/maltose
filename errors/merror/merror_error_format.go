package merror

import (
	"fmt"
	"io"
)

// Format 实现了 fmt.Formatter 接口，它可以格式化错误信息。
//
// 格式说明符：
//
//	%s: 错误信息
//	+v: 错误信息和堆栈信息
func (err *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		switch {
		case s.Flag('-'):
			if err.text != "" {
				_, _ = io.WriteString(s, err.text)
			} else {
				_, _ = io.WriteString(s, err.Error())
			}
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
