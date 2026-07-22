package mhttp

import (
	"bytes"
	"io"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graingo/maltose/os/mlog"
)

// LogMaxBodySize controls request/response body logging. Zero disables body logging.
var LogMaxBodySize = 0

// responseWriter is a custom http.ResponseWriter that captures the response body and status.
// It embeds gin.ResponseWriter to ensure full compatibility.
type responseWriter struct {
	gin.ResponseWriter
	body  *bytes.Buffer
	limit int
}

func (w *responseWriter) capture(data []byte) {
	if w.limit == 0 {
		return
	}
	if w.limit < 0 {
		_, _ = w.body.Write(data)
		return
	}
	remaining := w.limit + 1 - w.body.Len()
	if remaining <= 0 {
		return
	}
	if len(data) > remaining {
		data = data[:remaining]
	}
	_, _ = w.body.Write(data)
}

// Write writes the data to the connection as part of an HTTP reply.
// It writes to both the original writer and our buffer to capture the body.
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	if err == nil {
		w.capture(b[:n])
	}
	return n, err
}

// WriteString writes the string to the connection as part of an HTTP reply.
// It writes to both the original writer and our buffer to capture the body.
func (w *responseWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	if err == nil {
		w.capture([]byte(s[:n]))
	}
	return n, err
}

// MiddlewareLog is a middleware for logging HTTP requests in two steps:
// 1. Before the handler is executed ("started").
// 2. After the handler is completed ("finished").
// This allows for better observability, especially for hanging or panicking requests.
func MiddlewareLog() MiddlewareFunc {
	bodyLimit := LogMaxBodySize
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
		if bodyLimit != 0 && r.Request.Body != nil {
			reqBodyBytes, r.Request.Body = readBodyForLog(r.Request.Body, bodyLimit)
		}

		// Create a custom response writer to capture the response body and status.
		writer := &responseWriter{
			ResponseWriter: r.Writer,
			body:           &bytes.Buffer{},
			limit:          bodyLimit,
		}
		r.Writer = writer

		requestFields := mlog.Fields{
			mlog.String("ip", r.ClientIP()),
			mlog.String("method", r.Request.Method),
			mlog.String("path", r.Request.URL.Path),
		}
		if query := r.Request.URL.Query(); len(query) > 0 {
			keys := make([]string, 0, len(query))
			for key := range query {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			requestFields = append(requestFields, mlog.Any("query_keys", keys))
		}
		if len(reqBodyBytes) > 0 {
			requestFields = append(requestFields, mlog.String("request_body", getBodyString(reqBodyBytes, bodyLimit)))
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
		if bodyLimit != 0 && len(resBodyBytes) > 0 {
			finalFields = append(finalFields, mlog.String("response_body", getBodyString(resBodyBytes, bodyLimit)))
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

type replayReadCloser struct {
	io.Reader
	io.Closer
}

func readBodyForLog(body io.ReadCloser, limit int) ([]byte, io.ReadCloser) {
	if limit < 0 {
		data, err := io.ReadAll(body)
		if err != nil {
			return nil, body
		}
		_ = body.Close()
		return data, io.NopCloser(bytes.NewReader(data))
	}
	data, err := io.ReadAll(io.LimitReader(body, int64(limit+1)))
	if err != nil {
		return nil, body
	}
	return data, &replayReadCloser{
		Reader: io.MultiReader(bytes.NewReader(data), body),
		Closer: body,
	}
}

// getBodyString safely converts a byte slice to a string for logging, with a size limit.
func getBodyString(body []byte, limit int) string {
	if len(body) == 0 {
		return ""
	}
	if limit < 0 { // no limit
		return string(body)
	}
	if len(body) > limit {
		return string(body[:limit]) + "..."
	}
	return string(body)
}
