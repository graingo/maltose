package mlog

import (
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
	if entry.Context == nil {
		return nil
	}
	if traceID := mtrace.GetTraceID(entry.Context); traceID != "" {
		entry.Data["trace_id"] = traceID
	}
	if spanID := mtrace.GetSpanID(entry.Context); spanID != "" {
		entry.Data["span_id"] = spanID
	}
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
