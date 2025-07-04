package mclient

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mtrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentName                        = "github.com/graingo/maltose/net/mclient"
	tracingEventHttpRequest               = "http.request"
	tracingEventHttpRequestUrl            = "http.url"
	tracingEventHttpHeaders               = "http.headers"
	tracingEventHttpBaggage               = "http.baggage"
	tracingEventHttpResponse              = "http.response"
	tracingMiddlewareHandled   contextKey = "tracing-middleware-handled"
	version                               = maltose.VERSION
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
					err = merror.Newf("client panic: %v", r)
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
			ctx := req.Request.Context()

			// create tracer
			tr := otel.GetTracerProvider().Tracer(
				instrumentName,
				trace.WithInstrumentationVersion(version),
			)

			// build span name
			spanName := "HTTP " + req.Request.Method

			// Create span
			ctx, span := tr.Start(
				ctx,
				spanName,
				trace.WithSpanKind(trace.SpanKindClient),
			)
			defer span.End()

			// set span attributes
			span.SetAttributes(
				attribute.String(mtrace.AttributeHTTPMethod, req.Request.Method),
				attribute.String(mtrace.AttributeHTTPUrl, req.Request.URL.String()),
				attribute.String(mtrace.AttributeHTTPHost, getHost(req.Request.URL)),
				attribute.String(mtrace.AttributeHTTPScheme, getSchema(req.Request.URL)),
				attribute.String(mtrace.AttributeHTTPFlavor, getProtocolVersion(req.Request.Proto)),
				attribute.String(mtrace.AttributeHTTPTarget, getPath(req.Request.URL)),
				attribute.String(mtrace.AttributeHTTPUserAgent, req.Request.UserAgent()),
			)

			// Inject context into outgoing headers
			otel.GetTextMapPropagator().Inject(
				ctx,
				propagation.HeaderCarrier(req.Request.Header),
			)

			// Update request with tracing context
			req.Request = req.Request.WithContext(ctx)

			// Execute next middleware
			resp, err := next(req)

			// handle response and error
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else if resp != nil {
				span.SetAttributes(attribute.Int(mtrace.AttributeHTTPStatusCode, resp.StatusCode))
				statusCode, _ := httpStatusCodeToSpanStatus(resp.StatusCode)
				span.SetStatus(statusCode, "")
			}

			return resp, err
		}
	}
}

// httpStatusCodeToSpanStatus converts an HTTP status code to a span status code.
// It returns the span status code and a description.
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
