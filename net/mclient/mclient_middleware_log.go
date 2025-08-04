package mclient

import (
	"bytes"
	"io"
	"time"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/os/mlog"
)

var LogMaxBodySize = -1

// MiddlewareLog creates a middleware that logs request and response details in two steps:
// 1. Before the request is sent ("started").
// 2. After the request is completed ("finished" or "error").
// This allows for better observability, especially for hanging requests.
func MiddlewareLog(logger *mlog.Logger) MiddlewareFunc {
	if logger == nil {
		return func(next HandlerFunc) HandlerFunc {
			return next
		}
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			ctx := req.Context()
			l := logger.With(mlog.String(maltose.COMPONENT, "mclient"))

			// --- Step 1: Log request start ---

			var reqBodyBytes []byte
			if req.Body != nil {
				// Safely read request body for logging and then restore it
				reqBodyBytes, _ = io.ReadAll(req.Body)
				req.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			}

			requestFields := mlog.Fields{
				mlog.String("method", req.Request.Method),
				mlog.String("url", req.Request.URL.String()),
			}
			if len(reqBodyBytes) > 0 {
				requestFields = append(requestFields, mlog.String("request_body", getBodyString(reqBodyBytes, LogMaxBodySize)))
			}

			l.Infow(ctx, "http client request started", requestFields...)

			// --- Step 2: Execute request and log completion ---

			start := time.Now()
			resp, err := next(req)
			duration := time.Since(start)

			// The final log should contain all information for context.
			// Start with the initial request fields.
			finalFields := append(requestFields, mlog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6))

			if err != nil {
				// Handle network or other errors before getting a response
				finalFields = append(finalFields, mlog.Err(err))
				l.Errorw(ctx, err, "http client request error", finalFields...)
				return resp, err
			}

			// If we got a response, add its details to the log
			finalFields = append(finalFields, mlog.Int("status", resp.StatusCode))
			if resp.Body != nil {
				bodyBytes, _ := io.ReadAll(resp.Body)
				resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for next handler
				if len(bodyBytes) > 0 {
					finalFields = append(finalFields, mlog.String("response_body", getBodyString(bodyBytes, LogMaxBodySize)))
				}
			}

			if resp.StatusCode >= 400 {
				l.Warnw(ctx, "http client request finished with error status", finalFields...)
			} else {
				l.Infow(ctx, "http client request finished", finalFields...)
			}

			return resp, nil
		}
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
