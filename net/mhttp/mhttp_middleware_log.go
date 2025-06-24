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
		start := time.Now()

		// Safely read and log the request body
		var bodyBytes []byte
		if r.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Request.Body)
			r.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body
		}

		// Create a response writer to capture the response
		writer := &responseWriter{
			ResponseWriter: r.Writer,
			body:           &bytes.Buffer{},
		}
		r.Writer = writer

		// Log request details
		logFields := mlog.Fields{
			"ip":     r.ClientIP(),
			"method": r.Request.Method,
			"path":   r.Request.URL.Path,
		}
		if len(bodyBytes) > 0 {
			logFields["body"] = getBodyString(bodyBytes, 512)
		}
		if raw := r.Request.URL.RawQuery; raw != "" {
			logFields["query"] = raw
		}
		r.Logger().Info(r.Request.Context(), "incoming request", logFields)

		// Execute next middleware
		r.Next()

		// Log response details
		latency := time.Since(start)
		status := writer.Status()
		resBodyBytes := writer.body.Bytes()

		respLogFields := mlog.Fields{
			"status":  status,
			"latency": latency.String(),
			"body":    getBodyString(resBodyBytes, 512),
		}

		if len(r.Errors) > 0 {
			respLogFields["errors"] = r.Errors.String()
			r.Logger().Error(r.Request.Context(), "request finished", respLogFields)
		} else {
			r.Logger().Info(r.Request.Context(), "request finished", respLogFields)
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
