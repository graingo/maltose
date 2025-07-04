package mclient

import (
	"bytes"
	"io"
	"time"

	"github.com/graingo/maltose"
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
			l := logger.With(mlog.String(maltose.COMPONENT, "mclient"))

			var reqBodyBytes []byte
			if req.Body != nil {
				reqBodyBytes, _ = io.ReadAll(req.Body)
				req.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			}

			// Execute request
			resp, err := next(req)
			duration := time.Since(start)

			fields := buildLogFields(req, resp, duration, reqBodyBytes)

			if err != nil {
				l.Errorw(ctx, err, "http client request error", fields...)
				return resp, err
			}

			if resp.StatusCode >= 400 {
				l.Errorw(ctx, err, "http client request finished with error status", fields...)
			} else {
				l.Infow(ctx, "http client request finished", fields...)
			}

			return resp, nil
		}
	}
}

// buildLogFields extracts and builds a map of fields for logging.
func buildLogFields(r *Request, resp *Response, duration time.Duration, reqBodyBytes []byte) mlog.Fields {
	fields := mlog.Fields{
		mlog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
		mlog.String("method", r.Request.Method),
	}

	if r.Request != nil && r.Request.URL != nil {
		fields = append(fields, mlog.String("url", r.Request.URL.String()))
	} else {
		fields = append(fields, mlog.String("url", "<no url>"))
	}

	// Use the pre-read byte slice instead of trying to read the consumed stream.
	if len(reqBodyBytes) > 0 {
		bodyStr := string(reqBodyBytes)
		if len(bodyStr) > maxBodySize {
			bodyStr = bodyStr[:maxBodySize] + "..."
		}
		fields = append(fields, mlog.String("request_body", bodyStr))
	}

	if resp != nil {
		fields = append(fields, mlog.Int("status", resp.StatusCode))
		// Safely read and log the response body
		if resp.Body != nil {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body
			bodyStr := string(bodyBytes)
			if len(bodyStr) > maxBodySize {
				bodyStr = bodyStr[:maxBodySize] + "..."
			}
			fields = append(fields, mlog.String("response_body", bodyStr))
		}
	}

	return fields
}
