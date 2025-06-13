package mtrace

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/net/mtrace/internal/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Semantic conventions for HTTP.
const (
	AttributeHTTPMethod       = "http.method"
	AttributeHTTPUrl          = "http.url"
	AttributeHTTPTarget       = "http.target"
	AttributeHTTPHost         = "http.host"
	AttributeHTTPScheme       = "http.scheme"
	AttributeHTTPStatusCode   = "http.status_code"
	AttributeHTTPStatusText   = "http.status_text"
	AttributeHTTPFlavor       = "http.flavor"
	AttributeHTTPUserAgent    = "http.user_agent"
	AttributeHTTPRequestSize  = "http.request_content_length"
	AttributeHTTPResponseSize = "http.response_content_length"
	AttributeHTTPRoute        = "http.route"
	AttributeHTTPClientIP     = "http.client_ip"
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

// IsUsingDefaultProvider checks if the default trace provider is used.
func IsUsingDefaultProvider() bool {
	_, ok := otel.GetTracerProvider().(*provider.TracerProvider)
	return ok
}

// CheckSetDefaultTextMapPropagator checks if the default TextMapPropagator is set.
func CheckSetDefaultTextMapPropagator() {
	p := otel.GetTextMapPropagator()
	if len(p.Fields()) == 0 {
		otel.SetTextMapPropagator(GetDefaultTextMapPropagator())
	}
}

// GetDefaultTextMapPropagator returns the default TextMapPropagator for context propagation.
func GetDefaultTextMapPropagator() propagation.TextMapPropagator {
	return defaultTextMapPropagator
}

// GetTraceID gets the trace id from the context.
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

// GetSpanID gets the span id from the context.
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

// SetBaggageValue is a convenient function for adding a key-value pair to baggage.
func SetBaggageValue(ctx context.Context, key string, value any) context.Context {
	return NewBaggage(ctx).SetValue(key, value)
}

// SetBaggageMap is a convenient function for adding multiple key-value pairs to baggage.
func SetBaggageMap(ctx context.Context, data map[string]any) context.Context {
	return NewBaggage(ctx).SetMap(data)
}

// GetBaggageMap gets and returns the map of baggage values.
func GetBaggageMap(ctx context.Context) map[string]any {
	return NewBaggage(ctx).GetMap()
}

// GetBaggageVar gets and returns the value of the specified key from baggage.
func GetBaggageVar(ctx context.Context, key string) *mvar.Var {
	return NewBaggage(ctx).GetVar(key)
}

// WithUUID injects a custom UUID as trace id into the context.
func WithUUID(ctx context.Context, uuid string) (context.Context, error) {
	return WithTraceID(ctx, strings.Replace(uuid, "-", "", -1))
}

// WithTraceID injects a custom trace id into the context.
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
