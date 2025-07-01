package mlogz

import (
	"context"

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

func (h *traceHook) Fire(ctx context.Context, msg string, attrs []Attr) (string, []Attr) {
	if traceID := mtrace.GetTraceID(ctx); traceID != "" {
		attrs = append(attrs, String("trace_id", traceID))
	}
	if spanID := mtrace.GetSpanID(ctx); spanID != "" {
		attrs = append(attrs, String("span_id", spanID))
	}
	return msg, attrs
}

// ctxHook is a hook that extracts values from context.
type ctxHook struct {
	keys map[string]any
}

func (h *ctxHook) Name() string { return ctxHookName }

func (h *ctxHook) Levels() []Level { return AllLevels() }

func (h *ctxHook) Fire(ctx context.Context, msg string, attrs []Attr) (string, []Attr) {
	if ctx == nil {
		return msg, attrs
	}
	for attrKey, ctxKey := range h.keys {
		if value := ctx.Value(ctxKey); value != nil {
			attrs = append(attrs, Any(attrKey, value))
		}
	}
	return msg, attrs
}
