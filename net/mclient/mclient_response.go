package mclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/graingo/maltose/internal/intlog"
)

// Response is the struct for client request response.
type Response struct {
	*http.Response                   // Response is the underlying http.Response object of certain request.
	request        *Request          // The client request object that generated this response, including the underlying http.Request
	cookies        map[string]string // Response cookies, which are only parsed once.
}

// GetRequest returns the underlying http.Request.
func (r *Response) GetRequest() *http.Request {
	if r.request != nil && r.request.Request != nil {
		return r.request.Request
	}
	return nil
}

// initCookie initializes the cookie map attribute of Response.
func (r *Response) initCookie() {
	if r.cookies == nil {
		r.cookies = make(map[string]string)
		// Response might be nil.
		if r.Response != nil {
			for _, v := range r.Cookies() {
				r.cookies[v.Name] = v.Value
			}
		}
	}
}

// GetCookie retrieves and returns the cookie value of specified `key`.
func (r *Response) GetCookie(key string) string {
	r.initCookie()
	return r.cookies[key]
}

// GetCookieMap retrieves and returns a copy of current cookie values map.
func (r *Response) GetCookieMap() map[string]string {
	r.initCookie()
	m := make(map[string]string, len(r.cookies))
	for k, v := range r.cookies {
		m[k] = v
	}
	return m
}

// ReadAll retrieves and returns the response content as []byte.
func (r *Response) ReadAll() []byte {
	// Response might be nil.
	if r == nil || r.Response == nil {
		return []byte{}
	}
	body, err := io.ReadAll(r.Response.Body)
	if err != nil {
		// This logs error internally without interrupting execution flow
		intlog.Error(r.request.Context(), "ReadAll error:", err)
		return []byte{}
	}
	// Reset Body for multiple reads
	r.SetBodyContent(body)
	return body
}

// ReadAllString retrieves and returns the response content as string.
func (r *Response) ReadAllString() string {
	return string(r.ReadAll())
}

// Parse parses the response JSON to the given struct pointer.
// It returns error if response is nil or parsing fails.
func (r *Response) Parse(v interface{}) error {
	if r == nil || r.Response == nil {
		return fmt.Errorf("nil response")
	}

	// Check content type (optional, but recommended)
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !bytes.Contains([]byte(contentType), []byte("application/json")) {
		if r.request != nil {
			intlog.Printf(r.request.Context(), "Warning: Content-Type is not application/json: %s", contentType)
		}
	}

	// Read response body
	body := r.ReadAll()
	if len(body) == 0 {
		return fmt.Errorf("empty response body")
	}

	// Parse JSON
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("JSON unmarshal error: %w", err)
	}

	return nil
}

// IsSuccess returns whether the response status code is in the 2xx range,
// indicating that the request was successfully received, understood, and accepted.
func (r *Response) IsSuccess() bool {
	if r == nil || r.Response == nil {
		return false
	}
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// SetBodyContent overwrites response content with custom one.
func (r *Response) SetBodyContent(content []byte) {
	buffer := bytes.NewBuffer(content)
	r.Body = io.NopCloser(buffer)
	r.ContentLength = int64(buffer.Len())
}

// Close closes the response when it will never be used.
func (r *Response) Close() error {
	if r == nil || r.Response == nil {
		return nil
	}
	return r.Response.Body.Close()
}
