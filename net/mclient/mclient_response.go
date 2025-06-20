package mclient

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/internal/intlog"
)

// Response is the struct for client request response.
type Response struct {
	*http.Response                   // Response is the underlying http.Response object of certain request.
	cookies        map[string]string // Response cookies, which are only parsed once.
	result         any               // Result object for successful response.
	errorResult    any               // Error result object for error response.
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

// GetCookies retrieves and returns all cookie values.
func (r *Response) GetCookies() map[string]string {
	r.initCookie()
	return r.cookies
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
		intlog.Error(r.Request.Context(), "ReadAll error:", err)
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

// Parse parses the response body into the given result.
func (r *Response) Parse(result interface{}) error {
	if r.Response == nil || r.Response.Body == nil {
		return errors.New("response or response body is nil")
	}
	defer r.Response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(r.Response.Body)
	if err != nil {
		return err
	}

	// Reset Body for multiple reads
	r.SetBodyContent(body)

	// Attempt to parse the response body
	mediaType := r.Header.Get("Content-Type")
	if idx := strings.Index(mediaType, ";"); idx != -1 {
		mediaType = mediaType[:idx]
	}
	mediaType = strings.TrimSpace(strings.ToLower(mediaType)) // Normalize to lower case and trim spaces
	switch mediaType {
	case "application/json":
		return json.Unmarshal(body, result)
	case "application/xml", "text/xml":
		return xml.Unmarshal(body, result)
	case "text/plain":
	default:
		resultPtr, ok := result.(*string)
		if !ok {
			return merror.Newf("mclient: text/plain content type requires a *string to unmarshal into, got %T", result)
		}
		*resultPtr = string(body)
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

// SetResult sets the result object for successful response.
func (r *Response) SetResult(result interface{}) {
	r.result = result
}

// SetError sets the error result object for error response.
func (r *Response) SetError(err interface{}) {
	r.errorResult = err
}

// parseResponse parses the response based on status code.
func (r *Response) parseResponse() error {
	if r.Response == nil {
		return errors.New("response is nil")
	}

	intlog.Printf(r.Request.Context(), "parseResponse called, status code: %d, result: %v, errorResult: %v",
		r.StatusCode, r.result != nil, r.errorResult != nil)

	if r.StatusCode >= 200 && r.StatusCode < 300 {
		// Success response - parse into result if provided
		if r.result != nil {
			intlog.Printf(r.Request.Context(), "Parsing success response into result")
			return r.Parse(r.result)
		}
	} else {
		// Error response - parse into errorResult if provided
		if r.errorResult != nil {
			intlog.Printf(r.Request.Context(), "Parsing error response into errorResult")
			return r.Parse(r.errorResult)
		}
	}

	return nil
}

// GetResult returns the result object.
func (r *Response) GetResult() any {
	return r.result
}

// GetError returns the error result object.
func (r *Response) GetError() any {
	return r.errorResult
}
