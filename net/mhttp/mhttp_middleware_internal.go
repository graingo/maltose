package mhttp

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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
	instrumentName = "github.com/graingo/maltose/net/mhttp"
	version        = maltose.VERSION
)

// internalMiddlewareDefaultResponse internal default response processing middleware
func internalMiddlewareDefaultResponse() MiddlewareFunc {
	return func(r *Request) {
		r.Next()

		// if response has been written by other middleware, skip
		if r.Writer.Written() {
			return
		}

		// handle error case - only handle unstructured errors as fallback
		// Let user middleware handle structured errors with error codes
		if len(r.Errors) > 0 {
			err := r.Errors.Last().Err
			code := merror.Code(err)
			if code == mcode.CodeNil {
				r.String(500, fmt.Sprintf("Error: %s", err.Error()))
			} else {
				r.String(codeToHTTPStatus(code), code.Message())
			}
			return
		}

		// handle response from handler or other middleware
		if res := r.GetHandlerResponse(); res != nil {
			switch v := res.(type) {
			case string:
				r.String(200, v)
			case []byte:
				r.String(200, string(v))
			default:
				r.JSON(200, res)
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
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if errors.As(ne, &se) {
						seStr := strings.ToLower(se.Error())
						if strings.Contains(seStr, "broken pipe") ||
							strings.Contains(seStr, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				merr := merror.NewCodef(mcode.CodeInternalPanic, "Panic recovered: %s", err)

				if brokenPipe {
					// If the connection is dead, we can't write a status to it.
					// Just log the error and abort
					r.Logger().Warnf(r.Request.Context(), "Connection broken: %s", err)
					r.Error(merr)
					r.Abort()
				} else {
					// record error log for normal panic
					r.Logger().Errorf(r.Request.Context(), merr, "Panic recovered")
					// call panic handler
					if r.server.panicHandler != nil {
						r.server.panicHandler(r, merr)
					}
				}
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
		// Skip health check
		if r.Request.URL.Path == r.server.config.HealthCheck {
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

		// Inject the updated context into both Request objects
		r.Request = r.Request.WithContext(ctx)

		// process request
		r.Next()

		// set span attributes
		span.SetAttributes(
			attribute.String(mtrace.AttributeHTTPMethod, r.Request.Method),
			attribute.String(mtrace.AttributeHTTPUrl, r.Request.URL.String()),
			attribute.String(mtrace.AttributeHTTPHost, r.Request.Host),
			attribute.String(mtrace.AttributeHTTPScheme, getSchema(r)),
			attribute.String(mtrace.AttributeHTTPFlavor, r.Request.Proto),
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
