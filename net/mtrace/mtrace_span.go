package mtrace

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Span wraps trace.Span to provide compatibility and extensions.
type Span struct {
	trace.Span
}

// NewSpan creates a span using the default tracer.
func NewSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	ctx, span := NewTracer().Start(ctx, spanName, opts...)
	return ctx, &Span{
		Span: span,
	}
}
