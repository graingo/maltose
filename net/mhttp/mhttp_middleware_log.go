package mhttp

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graingo/maltose/os/mlog"
)

// responseWriter is a custom http.ResponseWriter that captures the response body and status.
// It embeds gin.ResponseWriter to ensure full compatibility.
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write writes the data to the connection as part of an HTTP reply.
// It writes to both the original writer and our buffer to capture the body.
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err == nil {
		w.body.Write(b[:n])
	}
	return n, err
}

// WriteString writes the string to the connection as part of an HTTP reply.
// It writes to both the original writer and our buffer to capture the body.
func (w *responseWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	if err == nil {
		w.body.WriteString(s[:n])
	}
	return n, err
}

// MiddlewareLog is a middleware for logging HTTP requests in two steps:
// 1. Before the handler is executed ("started").
// 2. After the handler is completed ("finished").
// This allows for better observability, especially for hanging or panicking requests.
func MiddlewareLog() MiddlewareFunc {
	return func(r *Request) {
		// Skip health check
		if r.Request.URL.Path == r.server.config.HealthCheck {
			r.Next()
			return
		}

		// --- Step 1: Log request start ---

		start := time.Now()

		// Safely read and capture the request body for logging, then restore it.
		var reqBodyBytes []byte
		if r.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(r.Request.Body)
			r.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		// Create a custom response writer to capture the response body and status.
		writer := &responseWriter{
			ResponseWriter: r.Writer,
			body:           &bytes.Buffer{},
		}
		r.Writer = writer

		requestFields := mlog.Fields{
			mlog.String("ip", r.ClientIP()),
			mlog.String("method", r.Request.Method),
			mlog.String("path", r.Request.URL.Path),
		}
		if raw := r.Request.URL.RawQuery; raw != "" {
			requestFields = append(requestFields, mlog.String("query", raw))
		}
		if len(reqBodyBytes) > 0 {
			requestFields = append(requestFields, mlog.String("request_body", getBodyString(reqBodyBytes, 512)))
		}

		r.Logger().Infow(r.Request.Context(), "http server request started", requestFields...)

		// --- Step 2: Execute handler and log completion ---

		r.Next()

		duration := time.Since(start)
		status := writer.Status()
		resBodyBytes := writer.body.Bytes()

		msg := "http server request finished"

		// The final log should contain all information for context.
		// Start with the initial request fields and add response details.
		finalFields := append(requestFields,
			mlog.Int("status", status),
			mlog.Float64("latency_ms", float64(duration.Nanoseconds())/1e6),
		)
		if len(resBodyBytes) > 0 {
			finalFields = append(finalFields, mlog.String("response_body", getBodyString(resBodyBytes, 512)))
		}

		// Decide log level based on errors or status code
		if len(r.Errors) > 0 {
			msg += " with errors"
			// Log with the actual error from the context
			r.Logger().Errorw(r.Request.Context(), r.Errors.Last().Err, msg, finalFields...)
		} else if status >= 400 {
			msg += " with warning status"
			r.Logger().Warnw(r.Request.Context(), msg, finalFields...)
		} else {
			r.Logger().Infow(r.Request.Context(), msg, finalFields...)
		}
	}
}

// getBodyString safely converts a byte slice to a string for logging, with a size limit.
func getBodyString(body []byte, limit int) string {
	if len(body) == 0 {
		return ""
	}
	if len(body) > limit {
		return string(body[:limit]) + "..."
	}
	return string(body)
}
