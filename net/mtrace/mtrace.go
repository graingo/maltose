package mtrace

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/savorelle/maltose/container/mvar"
	"github.com/savorelle/maltose/net/mtrace/internal/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	defaultTextMapPropagator = propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
)

func init() {
	otel.SetTracerProvider(provider.New())
	CheckSetDefaultTextMapPropagator()
}

// IsUsingDefaultProvider 检查是否使用默认的 trace provider
func IsUsingDefaultProvider() bool {
	_, ok := otel.GetTracerProvider().(*provider.TracerProvider)
	return ok
}

// CheckSetDefaultTextMapPropagator 如果之前未设置，则设置默认的 TextMapPropagator
func CheckSetDefaultTextMapPropagator() {
	p := otel.GetTextMapPropagator()
	if len(p.Fields()) == 0 {
		otel.SetTextMapPropagator(GetDefaultTextMapPropagator())
	}
}

// GetDefaultTextMapPropagator 返回用于对等体间上下文传播的默认传播器
func GetDefaultTextMapPropagator() propagation.TextMapPropagator {
	return defaultTextMapPropagator
}

// GetTraceID 从上下文中获取 TraceId
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID := trace.SpanContextFromContext(ctx).TraceID()
	if traceID.IsValid() {
		return traceID.String()
	}
	return ""
}

// GetSpanID 从上下文中获取 SpanId
func GetSpanID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	spanID := trace.SpanContextFromContext(ctx).SpanID()
	if spanID.IsValid() {
		return spanID.String()
	}
	return ""
}

// SetBaggageValue 是一个便捷函数，用于向 baggage 添加一个键值对
func SetBaggageValue(ctx context.Context, key string, value any) context.Context {
	return NewBaggage(ctx).SetValue(key, value)
}

// SetBaggageMap 是一个便捷函数，用于向 baggage 添加多个键值对
func SetBaggageMap(ctx context.Context, data map[string]any) context.Context {
	return NewBaggage(ctx).SetMap(data)
}

// GetBaggageMap 获取并返回 baggage 值的 map
func GetBaggageMap(ctx context.Context) map[string]any {
	return NewBaggage(ctx).GetMap()
}

// GetBaggageVar 从 baggage 中获取指定 key 的值并返回 *mvar.Var
func GetBaggageVar(ctx context.Context, key string) *mvar.Var {
	return NewBaggage(ctx).GetVar(key)
}

// WithUUID 将自定义的 UUID 作为 trace id 注入到上下文中
func WithUUID(ctx context.Context, uuid string) (context.Context, error) {
	return WithTraceID(ctx, strings.Replace(uuid, "-", "", -1))
}

// WithTraceID 将自定义的 trace id 注入到上下文中
func WithTraceID(ctx context.Context, traceID string) (context.Context, error) {
	generatedTraceID, err := trace.TraceIDFromHex(traceID)
	if err != nil {
		return ctx, errors.New(fmt.Sprintf(`invalid custom traceID "%s", a traceID string should be composed with [0-f] and fixed length 32`, traceID))
	}
	sc := trace.SpanContextFromContext(ctx)
	if !sc.HasTraceID() {
		var span trace.Span
		ctx, span = NewSpan(ctx, "gtrace.WithTraceID")
		defer span.End()
		sc = trace.SpanContextFromContext(ctx)
	}
	ctx = trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    generatedTraceID,
		SpanID:     sc.SpanID(),
		TraceFlags: sc.TraceFlags(),
		TraceState: sc.TraceState(),
		Remote:     sc.IsRemote(),
	}))
	return ctx, nil
}
