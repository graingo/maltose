package mctx

import (
	"context"

	"github.com/graingo/maltose/net/mtrace"
)

type (
	// Ctx is a short name alias for context.Context.
	Ctx = context.Context
)

// New creates and returns a context which contains context id.
// The created context has a new isolated trace id for tracing functionality.
func New() context.Context {
	return WithSpan(context.Background(), "mctx.New")
}

// WithSpan creates and returns a context containing span upon given parent context `ctx`.
func WithSpan(ctx context.Context, spanName string) context.Context {
	if spanName == "" {
		spanName = "mctx.WithSpan"
	}
	ctx, span := mtrace.NewSpan(ctx, spanName)
	// The API returns only a context, so callers have no span handle to end.
	// End the synthetic context span immediately to avoid retaining it forever.
	span.End()
	return ctx
}

// CtxID retrieves and returns the context id from context.
func CtxID(ctx context.Context) string {
	return mtrace.GetTraceID(ctx)
}
