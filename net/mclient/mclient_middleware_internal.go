package mclient

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/graingo/maltose/net/mtrace"
	"github.com/graingo/mconv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentName                        = "github.com/graingo/maltose/net/mclient.client"
	tracingEventHttpRequest               = "http.request"
	tracingEventHttpRequestUrl            = "http.request.url"
	tracingEventHttpHeaders               = "http.headers"
	tracingEventHttpBaggage               = "http.baggage"
	tracingEventHttpResponse              = "http.response"
	tracingMiddlewareHandled   contextKey = "TracingMiddlewareHandled"
)

type contextKey string

// internalMiddlewareRecovery internal error recovery middleware
func internalMiddlewareRecovery() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			var resp *Response
			var err error

			defer func() {
				if r := recover(); r != nil {
					// Handle panic
					err = fmt.Errorf("client panic: %v", r)
				}
			}()

			resp, err = next(req)
			return resp, err
		}
	}
}

// internalMiddlewareMetric internal metric collection middleware
func internalMiddlewareMetric() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			// Record start time
			startTime := time.Now()

			// Collect metrics before request
			handleMetricsBeforeRequest(req.Request)

			// Execute next middleware
			resp, err := next(req)

			// Collect metrics after request done
			if resp != nil {
				handleMetricsAfterRequestDone(req.Request, resp.Response, err, startTime)
			} else {
				handleMetricsAfterRequestDone(req.Request, nil, err, startTime)
			}

			return resp, err
		}
	}
}

// internalMiddlewareTrace client tracing middleware
func internalMiddlewareTrace() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			if req.Request == nil {
				return next(req)
			}

			ctx := req.Context()

			// Avoid duplicate processing
			if ctx.Value(tracingMiddlewareHandled) != nil {
				return next(req)
			}

			ctx = context.WithValue(ctx, tracingMiddlewareHandled, 1)

			// If using default provider, skip complex tracing
			if mtrace.IsUsingDefaultProvider() {
				req.Request = req.Request.WithContext(ctx)
				return next(req)
			}

			// Create tracer
			tr := otel.GetTracerProvider().Tracer(
				instrumentName,
				trace.WithInstrumentationVersion("v1.0.0"),
			)

			// Normalize operation name: HTTP {method} {url}
			urlPath := req.Request.URL.Path
			if urlPath == "" {
				urlPath = "/"
			}
			spanName := fmt.Sprintf("HTTP %s %s", req.Request.Method, urlPath)

			// Create span
			ctx, span := tr.Start(
				ctx,
				spanName,
				trace.WithSpanKind(trace.SpanKindClient),
			)
			defer span.End()

			// Add request event
			span.AddEvent(tracingEventHttpRequest, trace.WithAttributes(
				attribute.String(tracingEventHttpRequestUrl, req.Request.URL.String()),
				attribute.String(tracingEventHttpHeaders, mconv.ToString(mconv.ToStringMap(req.Request.Header))),
				attribute.String(tracingEventHttpBaggage, mconv.ToString(mtrace.GetBaggageMap(ctx))),
			))

			// Inject context into outgoing headers
			otel.GetTextMapPropagator().Inject(
				ctx,
				propagation.HeaderCarrier(req.Request.Header),
			)

			// Update request with tracing context
			req.Request = req.Request.WithContext(ctx)

			// Execute next middleware
			resp, err := next(req)

			// Handle error
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				return resp, err
			}

			// Add response event
			if resp != nil && resp.Response != nil {
				statusCode := resp.StatusCode
				span.AddEvent(tracingEventHttpResponse, trace.WithAttributes(
					attribute.Int("http.status_code", statusCode),
					attribute.Int("http.response_content_length", int(resp.ContentLength)),
					attribute.String(tracingEventHttpHeaders, fmt.Sprint(resp.Header)),
				))

				spanStatus, statusMsg := httpStatusCodeToSpanStatus(statusCode)
				span.SetStatus(spanStatus, statusMsg)
			}

			return resp, err
		}
	}
}

// httpStatusCodeToSpanStatus converts HTTP status code to span status
func httpStatusCodeToSpanStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 400 {
		return codes.Error, fmt.Sprintf("HTTP %d", code)
	}
	return codes.Ok, ""
}

// getHost extracts the host from URL
func getHost(u *url.URL) string {
	if u.Host != "" {
		return u.Host
	}
	return "unknown"
}

// getSchema extracts the schema from URL
func getSchema(u *url.URL) string {
	if u.Scheme != "" {
		return u.Scheme
	}
	return "http"
}

// getPath extracts the path from URL
func getPath(u *url.URL) string {
	if u.Path != "" {
		return u.Path
	}
	return "/"
}

// getProtocolVersion extracts protocol version
func getProtocolVersion(proto string) string {
	if proto != "" {
		return strings.ToLower(proto)
	}
	return "http/1.1"
}
