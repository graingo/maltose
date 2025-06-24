package mclient

import (
	"bytes"
	"io"
	"time"

	"github.com/graingo/maltose/os/mlog"
)

const maxBodySize = 512

// MiddlewareLog creates a middleware that logs request and response details
// using the provided logger.
func MiddlewareLog(logger *mlog.Logger) MiddlewareFunc {
	if logger == nil {
		return func(next HandlerFunc) HandlerFunc {
			return next
		}
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			start := time.Now()
			ctx := req.Context()

			// Add component to logger for this middleware instance
			l := logger.WithComponent("mclient")

			// Execute request
			resp, err := next(req)
			duration := time.Since(start)

			fields := buildLogFields(req, resp, duration)

			if err != nil {
				l.Errorf(ctx, "http client request error: %v", err, fields)
				return resp, err
			}

			if resp.StatusCode >= 400 {
				l.Error(ctx, "http client request finished with error status", fields)
			} else {
				l.Info(ctx, "http client request finished", fields)
			}

			return resp, nil
		}
	}
}

// buildLogFields extracts and builds a map of fields for logging.
func buildLogFields(r *Request, resp *Response, duration time.Duration) mlog.Fields {
	fields := mlog.Fields{
		"duration_ms": float64(duration.Nanoseconds()) / 1e6,
		"method":      r.Method,
		"url":         "<no url>",
	}

	if r.Request != nil && r.Request.URL != nil {
		fields["url"] = r.Request.URL.String()
	}

	// Safely read and log the request body
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body
		bodyStr := string(bodyBytes)
		if len(bodyStr) > maxBodySize {
			bodyStr = bodyStr[:maxBodySize] + "..."
		}
		fields["request_body"] = bodyStr
	}

	if resp != nil {
		fields["status"] = resp.StatusCode
		// Safely read and log the response body
		if resp.Body != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body
			bodyStr := string(bodyBytes)
			if len(bodyStr) > maxBodySize {
				bodyStr = bodyStr[:maxBodySize] + "..."
			}
			fields["response_body"] = bodyStr
		}
	}

	return fields
}
