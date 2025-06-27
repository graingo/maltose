package mlogs

import (
	"context"
	"log/slog"
)

type CtxHandler struct {
	slog.Handler
	keys []string
}

func NewCtxHandler(h slog.Handler, keys []string) *CtxHandler {
	return &CtxHandler{h, keys}
}

func (h *CtxHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, key := range h.keys {
		if val := ctx.Value(key); val != nil {
			r.AddAttrs(slog.Any(key, val))
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (h *CtxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CtxHandler{
		Handler: h.Handler.WithAttrs(attrs),
		keys:    h.keys,
	}
}

func (h *CtxHandler) WithGroup(name string) slog.Handler {
	return &CtxHandler{
		Handler: h.Handler.WithGroup(name),
		keys:    h.keys,
	}
}
