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

func (h *traceHook) Fire(ctx context.Context, msg string, fields []Field) (string, []Field) {
	if traceID := mtrace.GetTraceID(ctx); traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}
	if spanID := mtrace.GetSpanID(ctx); spanID != "" {
		fields = append(fields, String("span_id", spanID))
	}
	return msg, fields
}

// ctxHook is a hook that extracts values from context.
type ctxHook struct {
	keys map[string]any
}

func (h *ctxHook) Name() string { return ctxHookName }

func (h *ctxHook) Levels() []Level { return AllLevels() }

func (h *ctxHook) Fire(ctx context.Context, msg string, fields []Field) (string, []Field) {
	if ctx == nil {
		return msg, fields
	}
	for attrKey, ctxKey := range h.keys {
		if value := ctx.Value(ctxKey); value != nil {
			fields = append(fields, Any(attrKey, value))
		}
	}
	return msg, fields
}
