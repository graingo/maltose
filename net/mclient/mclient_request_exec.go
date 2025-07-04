package mclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/graingo/maltose/internal/intlog"
)

// Send performs a request with the chain style API.
// If the method is not specified, it defaults to GET.
func (r *Request) Send(url string) (*Response, error) {
	if r.Request == nil || r.Request.Method == "" {
		// Default to GET method if not specified
		return r.doRequest(r.Request.Context(), http.MethodGet, url)
	}

	return r.doRequest(r.Request.Context(), r.Request.Method, url)
}

// doRequest sends the request and returns the response.
// This is an internal method used by Do.
func (r *Request) doRequest(ctx context.Context, method string, urlPath string) (*Response, error) {
	var (
		err      error
		resp     *Response
		attempts = 0
	)

	// Start with at least one attempt (0 retries)
	maxAttempts := r.retryCount + 1
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for attempts < maxAttempts {
		attempts++

		// Create a new request for each attempt
		resp, err = r.attemptRequest(ctx, method, urlPath)

		// If an error occurred (like a timeout), resp might be nil.
		// We must handle the error case first before accessing resp.
		if err != nil {
			// If we shouldn't retry based on the error, or we've exhausted attempts, break.
			if !r.shouldRetry(nil, err) || attempts >= maxAttempts {
				break
			}
		} else {
			// If there's no error, check the response to see if we should retry.
			if !r.shouldRetry(resp.Response, nil) || attempts >= maxAttempts {
				break
			}
		}

		// Close the response before retry if it exists
		if resp != nil {
			resp.Close()
			resp = nil
		}

		// Log retry attempt
		if r.Request != nil && r.Request.Context() != nil {
			intlog.Printf(r.Request.Context(), "Retrying request (attempt %d/%d) after error: %v",
				attempts, maxAttempts, err)
		}

		// Wait before retry if interval is set
		if r.retryInterval > 0 {
			select {
			case <-time.After(r.retryInterval):
				// Continue after waiting
			case <-ctx.Done():
				// Context cancelled during wait
				return nil, ctx.Err()
			}
		}
	}

	if err != nil {
		return nil, err
	}

	// Parse response if needed
	if err := resp.parseResponse(); err != nil {
		resp.Close()
		return nil, err
	}

	return resp, nil
}

// attemptRequest makes a single attempt to execute the request
func (r *Request) attemptRequest(ctx context.Context, method string, urlPath string) (*Response, error) {
	var (
		req *http.Request
		err error
	)

	// Prepare the request URL
	fullURL := urlPath
	if r.client.config.BaseURL != "" && !strings.HasPrefix(urlPath, "http://") && !strings.HasPrefix(urlPath, "https://") {
		baseURL := r.client.config.BaseURL

		// Ensure there's a single slash between baseURL and urlPath
		if !strings.HasSuffix(baseURL, "/") && !strings.HasPrefix(urlPath, "/") {
			baseURL = baseURL + "/"
		} else if strings.HasSuffix(baseURL, "/") && strings.HasPrefix(urlPath, "/") {
			urlPath = urlPath[1:]
		}

		fullURL = baseURL + urlPath
	}

	// Process query parameters
	if len(r.queryParams) > 0 {
		if strings.Contains(fullURL, "?") {
			fullURL = fullURL + "&" + r.queryParams.Encode()
		} else {
			fullURL = fullURL + "?" + r.queryParams.Encode()
		}
	}

	// Process form parameters
	var body io.Reader
	if len(r.formParams) > 0 {
		// Prioritize form data
		body = strings.NewReader(r.formParams.Encode())
		if r.Request == nil {
			r.Request = &http.Request{
				Header: make(http.Header),
			}
		}
		r.ContentType("application/x-www-form-urlencoded")
	} else if r.Request != nil && r.Request.Body != nil {
		// For retries, we need to make body re-readable
		if bodyBytes, err := io.ReadAll(r.Request.Body); err == nil {
			r.Request.Body.Close()
			body = bytes.NewReader(bodyBytes)
			r.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		} else {
			// If we can't read the body, use it directly
			// Note: this might cause issues with retries
			body = r.Request.Body
		}
	}

	// Create the HTTP request
	req, err = http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// Set headers from the client config
	if r.client.config.Header != nil {
		for k, v := range r.client.config.Header {
			if len(v) > 0 {
				req.Header.Set(k, v[0])
			}
		}
	}

	// Set headers from the request
	if r.Request != nil && r.Request.Header != nil {
		for k, v := range r.Request.Header {
			if len(v) > 0 {
				req.Header.Set(k, v[0])
			}
		}
	}

	// Set the updated http.Request in our Request object
	r.Request = req

	// Create a Response placeholder that will be filled by middleware
	var response *Response

	// Prepare the middleware chain
	middlewares := append(r.client.middlewares, r.middlewares...)
	if len(middlewares) > 0 {
		// Base handler - direct HTTP client call without middleware
		handler := func(req *Request) (*Response, error) {
			// At this point, use the underlying http.Request
			httpResp, err := r.client.do(req.Request)
			if err != nil {
				return nil, err
			}

			// Create Response object
			return &Response{
				Response:    httpResp,
				result:      req.result,
				errorResult: req.errorResult,
			}, nil
		}

		// Apply middlewares in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}

		// Execute the middleware chain with our Request object
		response, err = handler(r)
	} else {
		// Direct request without middleware
		httpResp, err := r.client.do(req)
		if err != nil {
			return nil, err
		}

		// Create Response object
		response = &Response{
			Response:    httpResp,
			result:      r.result,
			errorResult: r.errorResult,
		}
	}

	// Handle errors
	if err != nil {
		return nil, err
	}

	// Set response to request
	r.SetResponse(response)

	return response, nil
}

// Do executes the request with retry logic.
func (r *Request) Do() (*Response, error) {
	var resp *Response
	var err error

	// Try the request up to retryCount + 1 times
	for attempt := 0; attempt <= r.retryCount; attempt++ {
		if attempt > 0 {
			// Calculate delay for this attempt
			delay := r.calculateRetryDelay(attempt)
			time.Sleep(delay)
		}

		// Execute the request
		resp, err = r.doRequest(r.ctx, r.Request.Method, r.Request.URL.String())

		// Check if we should retry
		if err == nil && resp != nil {
			if r.retryCondition == nil {
				// Default retry condition: retry on 5xx and 429
				if resp.StatusCode < 500 && resp.StatusCode != 429 {
					return resp, nil
				}
			} else if !r.retryCondition(resp.Response, nil) {
				return resp, nil
			}
		} else if r.retryCondition == nil {
			// Default retry condition: retry on network errors
			if err != nil {
				continue
			}
		} else if !r.retryCondition(nil, err) {
			return resp, err
		}

		// Close response body if we're going to retry
		if resp != nil {
			resp.Close()
		}
	}

	return resp, err
}
