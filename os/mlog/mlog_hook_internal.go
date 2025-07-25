package mlog

import (
	"github.com/graingo/maltose/net/mtrace"
)

// internal hooks
const (
	traceHookName = "trace_hook"
	ctxHookName   = "ctx_hook"
)

// traceHook is a hook that automatically adds TraceID.
type traceHook struct{}

func (h *traceHook) Name() string { return traceHookName }

func (h *traceHook) Levels() []Level { return AllLevels() }

func (h *traceHook) Fire(entry *Entry) {
	if traceID := mtrace.GetTraceID(entry.GetContext()); traceID != "" {
		entry.AddField(String("trace.id", traceID))
	}
	if spanID := mtrace.GetSpanID(entry.GetContext()); spanID != "" {
		entry.AddField(String("span.id", spanID))
	}
}

// ctxHook is a hook that extracts values from context.
type ctxHook struct {
	keys []string
}

func (h *ctxHook) Name() string { return ctxHookName }

func (h *ctxHook) Levels() []Level { return AllLevels() }

func (h *ctxHook) Fire(entry *Entry) {
	if entry.GetContext() == nil {
		return
	}
	for _, ctxKey := range h.keys {
		if value := entry.GetContext().Value(ctxKey); value != nil {
			entry.AddField(Any(ctxKey, value))
		}
	}
}
