package mhttp

import (
	"fmt"
	"time"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/errors/mcode"
	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mtrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentName                        = "github.com/graingo/maltose/net/mhttp"
	tracingEventHttpRequest               = "http.request"
	tracingEventHttpRequestUrl            = "http.url"
	tracingEventHttpHeaders               = "http.headers"
	tracingEventHttpBaggage               = "http.baggage"
	tracingEventHttpResponse              = "http.response"
	tracingMiddlewareHandled   contextKey = "tracing-middleware-handled"
	version                               = maltose.VERSION
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
				r.JSON(200, v)
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
				r.Logger().Errorf(r.Request.Context(), nil, fmt.Sprintf("Panic recovered: %s", err))
				// return 500 error
				r.SetHandlerResponse(merror.NewCodef(mcode.CodeInternalError, "Panic recovered: %s", err))
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
		// If the default trace provider is used, do nothing.
		if !mtrace.IsUsingMaltoseProvider() {
			r.Next()
			return
		}

		ctx := r.Request.Context()
		tr := otel.GetTracerProvider().Tracer(
			instrumentName,
			trace.WithInstrumentationVersion(version),
		)

		// extract context and baggage
		ctx = otel.GetTextMapPropagator().Extract(
			ctx,
			propagation.HeaderCarrier(r.Request.Header),
		)

		// set span name
		spanName := r.Request.URL.Path
		if spanName == "" {
			spanName = "HTTP " + r.Request.Method
		}

		// start a new span
		ctx, span := tr.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Inject the updated context into the request.
		r.Request = r.Request.WithContext(ctx)

		// process request
		r.Next()

		// set span attributes
		span.SetAttributes(
			attribute.String(mtrace.AttributeHTTPMethod, r.Request.Method),
			attribute.String(mtrace.AttributeHTTPUrl, r.Request.URL.String()),
			attribute.String(mtrace.AttributeHTTPHost, r.Request.Host),
			attribute.String(mtrace.AttributeHTTPScheme, r.Request.URL.Scheme),
			attribute.String(mtrace.AttributeHTTPFlavor, r.Request.Proto),
			attribute.String(mtrace.AttributeHTTPTarget, r.Request.URL.Path),
			attribute.String(mtrace.AttributeHTTPUserAgent, r.Request.UserAgent()),
			attribute.String(mtrace.AttributeHTTPRoute, r.FullPath()),
			attribute.Int(mtrace.AttributeHTTPStatusCode, r.Writer.Status()),
		)

		// set span status
		if err := r.Errors.Last(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			statusCode, _ := httpStatusCodeToSpanStatus(r.Writer.Status())
			span.SetStatus(statusCode, "")
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
