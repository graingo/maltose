package mclient

import (
	"bytes"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/graingo/maltose"
	"github.com/graingo/maltose/os/mlog"
)

// LogMaxBodySize controls request/response body logging. Zero disables body logging.
var LogMaxBodySize = 0

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
	bodyLimit := LogMaxBodySize

	return func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			ctx := req.Context()
			l := logger.With(mlog.String(maltose.COMPONENT, "mclient"))

			// --- Step 1: Log request start ---

			var reqBodyBytes []byte
			if bodyLimit != 0 && req.Body != nil {
				reqBodyBytes, req.Body = readBodyForLog(req.Body, bodyLimit)
			}

			requestFields := mlog.Fields{
				mlog.String("method", req.Request.Method),
				mlog.String("url", urlWithoutQuery(req.Request)),
			}
			if queryKeys := requestQueryKeys(req.Request); len(queryKeys) > 0 {
				requestFields = append(requestFields, mlog.Any("query_keys", queryKeys))
			}
			if len(reqBodyBytes) > 0 {
				requestFields = append(requestFields, mlog.String("request_body", getBodyString(reqBodyBytes, bodyLimit)))
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
			if bodyLimit != 0 && resp.Body != nil {
				bodyBytes, restoredBody := readBodyForLog(resp.Body, bodyLimit)
				resp.Body = restoredBody
				if len(bodyBytes) > 0 {
					finalFields = append(finalFields, mlog.String("response_body", getBodyString(bodyBytes, bodyLimit)))
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

func urlWithoutQuery(request *http.Request) string {
	if request == nil || request.URL == nil {
		return ""
	}
	urlCopy := *request.URL
	urlCopy.RawQuery = ""
	urlCopy.ForceQuery = false
	return urlCopy.String()
}

func requestQueryKeys(request *http.Request) []string {
	if request == nil || request.URL == nil {
		return nil
	}
	query := request.URL.Query()
	keys := make([]string, 0, len(query))
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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
