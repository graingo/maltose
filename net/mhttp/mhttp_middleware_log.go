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

// MiddlewareLog is a middleware for logging HTTP requests.
func MiddlewareLog() MiddlewareFunc {
	return func(r *Request) {

		// Skip health check
		if r.Request.URL.Path == r.server.config.HealthCheck {
			r.Next()
			return
		}

		start := time.Now()

		// Safely read and capture the request body
		var reqBodyBytes []byte
		if r.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(r.Request.Body)
			r.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes)) // Restore body
		}

		// Create a response writer to capture the response
		writer := &responseWriter{
			ResponseWriter: r.Writer,
			body:           &bytes.Buffer{},
		}
		r.Writer = writer

		// Execute next middleware
		r.Next()

		// Now we have all the information, let's build the log fields
		duration := time.Since(start)
		status := writer.Status()
		resBodyBytes := writer.body.Bytes()

		fields := mlog.Fields{
			mlog.String("ip", r.ClientIP()),
			mlog.String("method", r.Request.Method),
			mlog.String("path", r.Request.URL.Path),
			mlog.Int("status", status),
			mlog.Float64("latency_ms", float64(duration.Nanoseconds())/1e6),
		}

		if raw := r.Request.URL.RawQuery; raw != "" {
			fields = append(fields, mlog.String("query", raw))
		}
		if len(reqBodyBytes) > 0 {
			fields = append(fields, mlog.String("request_body", getBodyString(reqBodyBytes, 512)))
		}
		if len(resBodyBytes) > 0 {
			fields = append(fields, mlog.String("response_body", getBodyString(resBodyBytes, 512)))
		}

		logger := r.Logger()
		msg := "http server request finished"

		// Decide log level based on errors or status code
		if len(r.Errors) > 0 {
			msg += " with errors"
			logger.Errorw(r.Request.Context(), r.Errors[0], msg, fields...)
		} else if status >= 400 {
			msg += " with error status"
			logger.Errorw(r.Request.Context(), nil, msg, fields...)
		} else {
			logger.Infow(r.Request.Context(), msg, fields...)
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
