package mlogs

import (
	"context"
	"log/slog"

	"github.com/graingo/maltose/net/mtrace"
)

type TraceHandler struct {
	slog.Handler
}

func NewTraceHandler(h slog.Handler) *TraceHandler {
	return &TraceHandler{h}
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID := mtrace.GetTraceID(ctx); traceID != "" {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	if spanID := mtrace.GetSpanID(ctx); spanID != "" {
		r.AddAttrs(slog.String("span_id", spanID))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
}
