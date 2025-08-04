// Package intlog provides internal logging for Maltose development usage only.
package intlog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/trace"
)

const (
	stackFilterKey = "/internal/intlog"
)

var (
	// debug is a flag for internal logging.
	// It's disabled by default.
	// It can be enabled by setting environment "MALTOSE_DEBUG" to "true".
	debug = false
)

func init() {
	if v := os.Getenv("MALTOSE_DEBUG"); v != "" {
		if b, _ := strconv.ParseBool(v); b {
			debug = true
		}
	}
}

// SetDebug sets the debug level for the internal logger.
func SetDebug(d bool) {
	debug = d
}

// Print prints `v` with newline using fmt.Println.
// The parameter `v` can be multiple variables.
func Print(ctx context.Context, v ...interface{}) {
	if !debug {
		return
	}
	doPrint(ctx, fmt.Sprint(v...), false)
}

// Printf prints `v` with format `format` using fmt.Printf.
// The parameter `v` can be multiple variables.
func Printf(ctx context.Context, format string, v ...interface{}) {
	if !debug {
		return
	}
	doPrint(ctx, fmt.Sprintf(format, v...), false)
}

// Error prints `v` with newline using fmt.Println.
// The parameter `v` can be multiple variables.
func Error(ctx context.Context, v ...interface{}) {
	if !debug {
		return
	}
	doPrint(ctx, fmt.Sprint(v...), true)
}

// Errorf prints `v` with format `format` using fmt.Printf.
func Errorf(ctx context.Context, format string, v ...interface{}) {
	if !debug {
		return
	}
	doPrint(ctx, fmt.Sprintf(format, v...), true)
}

func doPrint(ctx context.Context, content string, stack bool) {
	if !debug {
		return
	}

	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(time.Now().Format("2006-01-02 15:04:05.000"))
	buffer.WriteString(" [INTE] ")
	buffer.WriteString(file())
	buffer.WriteString(" ")
	if s := traceIDStr(ctx); s != "" {
		buffer.WriteString(s + " ")
	}
	buffer.WriteString(content)
	buffer.WriteString("\n")

	if stack {
		buffer.WriteString("Caller Stack:\n")
		callerStack := getCallerStack()
		buffer.WriteString(callerStack)
	}

	fmt.Print(buffer.String())
}

// traceIDStr retrieves and returns the trace id string for logging output.
func traceIDStr(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	spanCtx := trace.SpanContextFromContext(ctx)
	if traceID := spanCtx.TraceID(); traceID.IsValid() {
		return "{" + traceID.String() + "}"
	}
	return ""
}

// file returns caller file name along with its line number.
func file() string {
	_, file, line, ok := runtime.Caller(3) // Skip doPrint, Error/Print and the actual caller
	if ok {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
	return "unknown:0"
}

// getCallerStack returns the stack trace excluding this package.
func getCallerStack() string {
	stackBuf := bytes.NewBuffer(nil)

	// Start from depth 3 to skip doPrint, Error/Print functions
	for i := 3; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if isFilteredStack(file) {
			continue
		}
		stackBuf.WriteString(fmt.Sprintf("    %s:%d\n", filepath.Base(file), line))
	}

	return stackBuf.String()
}

// isFilteredStack checks if the stack frame should be filtered out.
func isFilteredStack(file string) bool {
	return filepath.Base(file) == "intlog.go"
}
