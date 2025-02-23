package mtrace

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Span 包装了 trace.Span 以提供兼容性和扩展
type Span struct {
	trace.Span
}

// NewSpan 使用默认的 tracer 创建一个 span
func NewSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	ctx, span := NewTracer().Start(ctx, spanName, opts...)
	return ctx, &Span{
		Span: span,
	}
}
