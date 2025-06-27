package mclient

import "net/http"

// SetHeader sets a header key-value pair for the request.
// This is an alias of Header method for better chain API compatibility.
func (r *Request) SetHeader(key, value string) *Request {
	if r.Request == nil {
		r.Request = &http.Request{
			Header: make(http.Header),
		}
	}
	r.Request.Header.Set(key, value)
	return r
}

// SetHeaders sets multiple headers at once.
func (r *Request) SetHeaders(headers map[string]string) *Request {
	for k, v := range headers {
		r.SetHeader(k, v)
	}
	return r
}

// ContentType sets the Content-Type header for the request.
func (r *Request) ContentType(contentType string) *Request {
	return r.SetHeader("Content-Type", contentType)
}
