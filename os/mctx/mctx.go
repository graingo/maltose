package mctx

import (
	"context"

	"github.com/graingo/maltose/net/mtrace"
)

type (
	// Ctx is a short name alias for context.Context.
	Ctx = context.Context
	// StrKey is a type for warps basic type string as context key.
	StrKey string
)

// New creates and returns a context which contains context id.
// The created context has a new isolated trace id for tracing functionality.
func New() context.Context {
	ctx, span := mtrace.NewSpan(context.Background(), "mctx.New")
	// The span is created just for creating a new trace id.
	// It is not necessary to use this span, so we end it directly.
	span.End()
	return ctx
}

// CtxId retrieves and returns the context id from context.
func CtxId(ctx context.Context) string {
	return mtrace.GetTraceID(ctx)
}
