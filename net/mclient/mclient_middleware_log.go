package mclient

import (
	"bytes"
	"io"
	"time"

	"github.com/graingo/maltose/os/mlog"
)

// MiddlewareLog creates a middleware that logs request and response details
// using the provided logger.
func MiddlewareLog(logger mlog.ILogger) MiddlewareFunc {
	if logger == nil {
		return func(next HandlerFunc) HandlerFunc {
			return next
		}
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			start := time.Now()
			ctx := req.Context()

			// Log request details
			logRequest(logger, req)

			// Execute request
			resp, err := next(req)

			// Log response or error
			if err != nil {
				logger.Errorf(ctx, "Request Failed, Duration: %s, Error: %v", time.Since(start), err)
				return resp, err
			}

			logResponse(logger, req, resp, time.Since(start))

			return resp, nil
		}
	}
}

// logRequest logs the details of the request.
func logRequest(logger mlog.ILogger, r *Request) {
	ctx := r.Context()
	urlStr := "<no url>"
	if r.Request != nil && r.Request.URL != nil {
		urlStr = r.Request.URL.String()
	}

	// Safely read and log the request body
	var bodyBytes []byte
	if r.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Request.Body)
		// Restore the body so it can be read again
		r.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Limit body size for logging to avoid spam
	const maxBodySize = 512
	bodyStr := string(bodyBytes)
	if len(bodyStr) > maxBodySize {
		bodyStr = bodyStr[:maxBodySize] + "..."
	}

	logger.Info(ctx, "[HTTP Client] Sending Request", map[string]any{
		"method": r.Request.Method,
		"url":    urlStr,
		"body":   bodyStr,
	})
}

// logResponse logs the details of the response.
func logResponse(logger mlog.ILogger, req *Request, resp *Response, duration time.Duration) {
	if resp == nil {
		logger.Warn(req.Context(), "[HTTP Client] Request Finished", map[string]any{
			"duration": duration.String(),
			"error":    "response is nil",
		})
		return
	}
	ctx := resp.Request.Context()

	// Safely read and log the response body
	var bodyBytes []byte
	if resp.Body != nil {
		bodyBytes, _ = io.ReadAll(resp.Body)
		// Restore the body so it can be read again
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Limit body size for logging
	const maxBodySize = 512
	bodyStr := string(bodyBytes)
	if len(bodyStr) > maxBodySize {
		bodyStr = bodyStr[:maxBodySize] + "..."
	}

	logFields := map[string]any{
		"status":   resp.StatusCode,
		"duration": duration.String(),
		"body":     bodyStr,
	}

	if resp.StatusCode >= 400 {
		logger.Error(ctx, "[HTTP Client] Request Finished", logFields)
	} else {
		logger.Info(ctx, "[HTTP Client] Request Finished", logFields)
	}
}
