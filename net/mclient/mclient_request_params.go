package mclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

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
func (r *Request) data(data any) *Request {
	if r.Request == nil {
		r.Request = &http.Request{
			Header: make(http.Header),
		}
	}

	switch d := data.(type) {
	case string:
		r.Request.Body = io.NopCloser(strings.NewReader(d))
	case []byte:
		r.Request.Body = io.NopCloser(bytes.NewReader(d))
	case io.Reader:
		r.Request.Body = io.NopCloser(d)
	default:
		// Try JSON encoding for other types
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			// Log error but continue execution
			// Using request context if available, otherwise fallback to background context
			ctx := context.Background()
			if r.Request != nil && r.Request.Context() != nil {
				ctx = r.Request.Context()
			}
			intlog.Error(ctx, "JSON marshal failed:", err)
			return r
		}
		r.Request.Body = io.NopCloser(bytes.NewReader(jsonBytes))
		if r.Request.Header.Get("Content-Type") == "" {
			r.ContentType("application/json")
		}
	}
	return r
}
