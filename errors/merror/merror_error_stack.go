package merror

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// Stack 返回错误的堆栈信息
func (err *Error) Stack() string {
	if err == nil {
		return ""
	}

	var (
		buffer bytes.Buffer
		pcs    [maxStackDepth]uintptr
		n      = runtime.Callers(3, pcs[:]) // 跳过前3层调用栈
	)

	// 写入错误信息
	buffer.WriteString(fmt.Sprintf("error: %s\n", err.Error()))
	buffer.WriteString("stack:\n")

	// 获取调用栈信息
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()

		// 跳过标准库和运行时的调用
		if strings.HasPrefix(frame.File, runtime.GOROOT()) {
			if !more {
				break
			}
			continue
		}

		// 格式化堆栈信息
		buffer.WriteString(fmt.Sprintf("  %s\n    %s:%d\n",
			frame.Function,
			frame.File,
			frame.Line,
		))

		if !more {
			break
		}
	}

	return buffer.String()
}
