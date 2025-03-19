package mhttp

import (
	"context"
	"fmt"
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
	instrumentName                        = "github.com/graingo/maltose/net/mhttp.server"
	tracingEventHttpRequest               = "http.request"
	tracingEventHttpRequestUrl            = "http.request.url"
	tracingEventHttpHeaders               = "http.headers"
	tracingEventHttpBaggage               = "http.baggage"
	tracingEventHttpResponse              = "http.response"
	tracingMiddlewareHandled   contextKey = "TracingMiddlewareHandled"
)

// internalMiddlewareDefaultResponse internal default response processing middleware
func internalMiddlewareDefaultResponse() MiddlewareFunc {
	return func(r *Request) {
		r.Next()

		// if response has been written, skip
		if r.Writer.Written() {
			return
		}

		// handle error case
		if len(r.Errors) > 0 {
			err := r.Errors.Last().Err
			r.String(500, fmt.Sprintf("Error: %s", err.Error()))
			return
		}

		// get handler response
		if res := r.GetHandlerResponse(); res != nil {
			switch v := res.(type) {
			case string:
				r.String(200, v)
			case []byte:
				r.String(200, string(v))
			default:
				r.String(200, fmt.Sprintf("%v", v))
			}
			return
		}

		// if no response, return empty string
		r.String(200, "")
	}
}

// internalMiddlewareRecovery internal error recovery middleware
func internalMiddlewareRecovery() MiddlewareFunc {
	return func(r *Request) {
		defer func() {
			if err := recover(); err != nil {
				// record error log
				r.Logger().Errorf(r.Request.Context(), "Panic recovered: %v", err)
				// return 500 error
				r.String(500, "Internal Server Error")
			}
		}()
		r.Next()
	}
}

// internalMiddlewareMetric internal metric collection middleware
func internalMiddlewareMetric() MiddlewareFunc {
	return func(r *Request) {
		// record start time
		startTime := time.Now()

		// collect metrics before request
		r.server.handleMetricsBeforeRequest(r)

		// execute next middleware
		r.Next()

		// collect metrics after request done
		r.server.handleMetricsAfterRequestDone(r, startTime)
	}
}

// internalMiddlewareTrace returns a middleware for OpenTelemetry tracing
func internalMiddlewareTrace() MiddlewareFunc {
	return func(r *Request) {
		ctx := r.Request.Context()

		// avoid duplicate processing
		if ctx.Value(tracingMiddlewareHandled) != nil {
			r.Next()
			return
		}

		ctx = context.WithValue(ctx, tracingMiddlewareHandled, 1)

		// if using default provider, skip complex tracing
		if mtrace.IsUsingDefaultProvider() {
			r.Next()
			return
		}

		// create tracer
		tr := otel.GetTracerProvider().Tracer(
			instrumentName,
			trace.WithInstrumentationVersion("v1.0.0"),
		)

		// extract context and baggage
		ctx = otel.GetTextMapPropagator().Extract(
			ctx,
			propagation.HeaderCarrier(r.Request.Header),
		)

		// normalize operation name: HTTP {method} {route}
		spanName := fmt.Sprintf("HTTP %s %s", r.Request.Method, r.FullPath())
		if r.FullPath() == "" {
			spanName = fmt.Sprintf("HTTP %s %s", r.Request.Method, r.Request.URL.Path)
		}

		// create span
		ctx, span := tr.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// add request event
		span.AddEvent(tracingEventHttpRequest, trace.WithAttributes(
			attribute.String(tracingEventHttpRequestUrl, r.Request.URL.String()),
			attribute.String(tracingEventHttpHeaders, mconv.ToString(mconv.ToStringMap(r.Request.Header))),
			attribute.String(tracingEventHttpBaggage, mconv.ToString(mtrace.GetBaggageMap(ctx))),
		))

		// inject tracing context
		r.Request = r.Request.WithContext(ctx)

		// continue processing request
		r.Next()

		// handle error
		if len(r.Errors) > 0 {
			// collect all error information
			var errMsgs []string
			for _, err := range r.Errors {
				errMsgs = append(errMsgs, err.Error())
			}
			span.SetStatus(codes.Error, strings.Join(errMsgs, "; "))
		} else {
			span.SetStatus(httpStatusCodeToSpanStatus(r.Writer.Status()))
		}

		// add response event
		span.AddEvent(tracingEventHttpResponse, trace.WithAttributes(
			attribute.Int("http.status_code", r.Writer.Status()),
			attribute.Int("http.response_content_length", r.Writer.Size()),
			attribute.String(tracingEventHttpHeaders, fmt.Sprint(r.Writer.Header())),
		))
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
