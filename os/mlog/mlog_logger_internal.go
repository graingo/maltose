package mlog

import (
	"fmt"

	"github.com/graingo/maltose/net/mtrace"
)

// traceHook is a hook that automatically adds TraceID.
type traceHook struct{}

// Levels implements the Hook interface.
func (h *traceHook) Levels() []Level {
	return AllLevels()
}

// Fire implements the Hook interface.
func (h *traceHook) Fire(entry *Entry) error {
	fmt.Println("traceHook Fire", mtrace.GetTraceID(entry.Context))
	entry.Data["trace_id"] = mtrace.GetTraceID(entry.Context)
	entry.Data["span_id"] = mtrace.GetSpanID(entry.Context)
	return nil
}

// ctxHook is a hook that extracts values from context.
type ctxHook struct {
	keys []string
}

// Levels implements the Hook interface.
func (h *ctxHook) Levels() []Level {
	return AllLevels()
}

// Fire implements the Hook interface.
func (h *ctxHook) Fire(entry *Entry) error {
	ctx := entry.Context
	if ctx == nil {
		return nil
	}

	// Extract values from context for each key
	for _, key := range h.keys {
		if value := ctx.Value(key); value != nil {
			entry.Data[key] = value
		}
	}

	return nil
}
