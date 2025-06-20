package mtrace

import (
	"context"
	"crypto/rand"
	"strings"

	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mtrace/internal/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
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
	CheckSetDefaultTextMapPropagator()
}

// IsUsingMaltoseProvider checks if the trace provider is a Maltose provider.
func IsUsingMaltoseProvider() bool {
	_, ok := otel.GetTracerProvider().(*provider.TracerProvider)
	return ok
}

// SetProvider sets the global tracer provider.
func SetProvider(p trace.TracerProvider) {
	otel.SetTracerProvider(p)
}

// NewProvider creates a new Maltose TracerProvider with custom options.
// It uses the OpenTelemetry SDK's default IDGenerator (random IDs) by default,
// which can be overridden by providing a custom `sdktrace.WithIDGenerator` option.
func NewProvider(opts ...sdkTrace.TracerProviderOption) trace.TracerProvider {
	return provider.New(opts...)
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
		return ctx, merror.Newf(`invalid custom traceID "%s", a traceID string should be composed with [0-f] and fixed length 32`, traceID)
	}

	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		// If there is no SpanContext in the current context,
		// we create a new one with the given traceID and a new random spanID.
		var spanID trace.SpanID
		_, _ = rand.Read(spanID[:])
		sc = trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: generatedTraceID,
			SpanID:  spanID,
			Remote:  true, // As it is from a custom ID, we mark it as remote.
		})
	} else {
		// If there is a SpanContext, we only replace the traceID.
		sc = sc.WithTraceID(generatedTraceID)
	}

	return trace.ContextWithSpanContext(ctx, sc), nil
}
