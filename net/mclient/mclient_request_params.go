package mclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/graingo/maltose/internal/intlog"
)

// SetQuery sets a query parameter for the request.
func (r *Request) SetQuery(key, value string) *Request {
	r.queryParams.Set(key, value)
	return r
}

// SetQueryMap sets multiple query parameters from a map.
func (r *Request) SetQueryMap(params map[string]string) *Request {
	for k, v := range params {
		r.queryParams.Set(k, v)
	}
	return r
}

// SetForm sets a form parameter for the request.
func (r *Request) SetForm(key, value string) *Request {
	r.formParams.Set(key, value)
	return r
}

// SetFormMap sets multiple form parameters from a map.
func (r *Request) SetFormMap(params map[string]string) *Request {
	for k, v := range params {
		r.formParams.Set(k, v)
	}
	return r
}

// SetBody sets the request body.
func (r *Request) SetBody(body any) *Request {
	return r.data(body)
}

// Data sets the request data.
// It intelligently handles different data types and buffers the body
// into memory to ensure it's re-readable for retries.
func (r *Request) data(data any) *Request {
	if r.Request == nil {
		r.Request = &http.Request{
			Header: make(http.Header),
		}
	}

	var bodyBytes []byte
	isJSON := false

	switch d := data.(type) {
	case string:
		bodyBytes = []byte(d)
	case []byte:
		bodyBytes = d
	case io.Reader:
		// For a generic, non-seekable reader, we must buffer it all into memory
		// to support retries. This is a design trade-off for simplicity and reliability.
		b, err := io.ReadAll(d)
		if err != nil {
			ctx := context.Background()
			if r.Request != nil && r.Request.Context() != nil {
				ctx = r.Request.Context()
			}
			intlog.Errorf(ctx, "mclient: failed to read io.Reader body for retry buffering: %v", err)
			// As a fallback, use the original reader but retries with body will fail.
			r.Request.Body = io.NopCloser(d)
			r.Request.GetBody = nil
			return r
		}
		bodyBytes = b
	default:
		// Try JSON encoding for other types
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			ctx := context.Background()
			if r.Request != nil && r.Request.Context() != nil {
				ctx = r.Request.Context()
			}
			intlog.Errorf(ctx, "mclient: json.Marshal failed: %v", err)
			return r
		}
		bodyBytes = jsonBytes
		isJSON = true
	}

	// Set the body for the first request
	r.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	// Set the content length
	r.Request.ContentLength = int64(len(bodyBytes))
	// Provide GetBody for retries, which is the standard way http.Client
	// handles re-sending the body on redirects or retries.
	r.Request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}

	if isJSON && r.Request.Header.Get("Content-Type") == "" {
		r.ContentType("application/json")
	}

	return r
}
