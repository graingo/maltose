package mclient

import (
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

			// Log request
			var urlStr string
			if req.Request != nil && req.Request.URL != nil {
				urlStr = req.Request.URL.String()
			} else {
				urlStr = "<no url>"
			}
			logger.Infof(ctx, "Request: %s %s", req.Method, urlStr)

			// Execute request
			resp, err := next(req)

			// Log response or error
			if err != nil {
				logger.Errorf(ctx, "Request failed: %v, Duration: %v", err, time.Since(start))
				return resp, err
			}

			if resp != nil {
				logger.Infof(ctx, "Response: %d, Duration: %v", resp.StatusCode, time.Since(start))
			} else {
				logger.Infof(ctx, "Response: nil, Duration: %v", time.Since(start))
			}

			return resp, nil
		}
	}
}
