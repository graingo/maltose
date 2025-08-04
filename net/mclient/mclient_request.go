package mclient

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Request is the struct for client request.
type Request struct {
	*http.Request                                   // Request is the underlying http.Request object.
	client         *Client                          // The client that creates this request.
	response       *Response                        // The response object of this request.
	retryCount     int                              // Retry count for the request.
	retryInterval  time.Duration                    // Retry interval for the request.
	middlewares    []MiddlewareFunc                 // Middleware functions.
	queryParams    url.Values                       // Query parameters.
	formParams     url.Values                       // Form parameters.
	retryCondition func(*http.Response, error) bool // Retry condition.
	retryConfig    RetryConfig                      // Retry configuration.
	result         any                              // Result object for successful response.
	errorResult    any                              // Error result object for error response.
}

// GetResponse returns the response object of this request.
func (r *Request) GetResponse() *Response {
	return r.response
}

// SetResponse sets the response object for this request.
func (r *Request) SetResponse(resp *Response) {
	r.response = resp
}

// SetContext sets the context for the request.
// It creates a new underlying http.Request with the given context.
func (r *Request) SetContext(ctx context.Context) *Request {
	if ctx == nil {
		return r
	}
	if r.Request == nil {
		r.Request = &http.Request{
			Header: make(http.Header),
		}
	}
	r.Request = r.Request.WithContext(ctx)
	return r
}

// Method sets the HTTP method for the request.
func (r *Request) Method(method string) *Request {
	if r.Request == nil {
		r.Request = &http.Request{
			Header: make(http.Header),
		}
	}
	r.Request.Method = method
	return r
}

// URL sets the request URL.
func (r *Request) URL(url string) *Request {
	if r.Request == nil {
		r.Request = &http.Request{
			Header: make(http.Header),
		}
	}
	parsed, err := r.Request.URL.Parse(url)
	if err == nil {
		r.Request.URL = parsed
	}
	return r
}

// SetResult sets the result object for successful response.
func (r *Request) SetResult(result any) *Request {
	r.result = result
	return r
}

// SetError sets the error result object for error response.
func (r *Request) SetError(err any) *Request {
	r.errorResult = err
	return r
}

// GetRequest returns the *http.Request object.
func (r *Request) GetRequest() *http.Request {
	return r.Request
}
