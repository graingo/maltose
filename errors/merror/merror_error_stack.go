package merror

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// Stack returns the stack information of the error.
func (err *Error) Stack() string {
	if err == nil {
		return ""
	}

	var (
		buffer bytes.Buffer
		pcs    [maxStackDepth]uintptr
		n      = runtime.Callers(3, pcs[:]) // Skip the first 3 stack frames
	)

	// Write error information
	buffer.WriteString(fmt.Sprintf("error: %s\n", err.Error()))
	buffer.WriteString("stack:\n")

	// Get the stack information
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()

		// Skip the calls to the standard library and runtime
		if strings.HasPrefix(frame.File, runtime.GOROOT()) {
			if !more {
				break
			}
			continue
		}

		// Format the stack information
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
