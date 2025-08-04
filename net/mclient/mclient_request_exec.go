package mclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/internal/intlog"
)

// Send performs a request with the chain style API.
// If the method is not specified, it defaults to GET.
func (r *Request) Send(url string) (*Response, error) {
	// Ensure the underlying http.Request is not nil
	if r.Request == nil {
		r.Request = &http.Request{}
	}
	return r.doRequest(r.Context(), r.Request.Method, url)
}

// doRequest manages the request execution, including the retry loop.
// It orchestrates calls to attemptRequest and handles the delay between retries.
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

		// If an error occurred (like a timeout from a panic), resp might be nil.
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

// attemptRequest makes a single attempt to execute the request.
// It builds the http.Request, chains and executes middlewares, and returns the response.
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

	var body io.Reader
	// It's essential to create a new body reader for every attempt.
	if len(r.formParams) > 0 {
		// Form data is safe to be re-encoded every time.
		body = strings.NewReader(r.formParams.Encode())
		if r.Request == nil {
			r.Request = &http.Request{
				Header: make(http.Header),
			}
		}
		r.ContentType("application/x-www-form-urlencoded")
	} else if r.Request != nil && r.Request.GetBody != nil {
		// Use GetBody to create a new reader for the body.
		var getBodyErr error
		body, getBodyErr = r.Request.GetBody()
		if getBodyErr != nil {
			return nil, merror.Wrap(getBodyErr, "failed to get request body for retry")
		}
	} else if r.Request != nil && r.Request.Body != nil {
		// This is a fallback for a non-nil, but non-resettable body like a live stream.
		// Retries will fail if the body is consumed.
		body = r.Request.Body
	}

	// Create the HTTP request for this attempt
	req, err = http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, merror.Wrapf(err, "http.NewRequestWithContext failed for method:%s, url:%s", method, fullURL)
	}

	// Preserve the GetBody function from the original request object,
	// as NewRequestWithContext does not automatically handle this for all body types.
	var originalGetBody func() (io.ReadCloser, error)
	if r.Request != nil {
		originalGetBody = r.Request.GetBody
		// Copy headers from the original request object that might have been set before Send().
		req.Header = r.Request.Header.Clone()
	}

	// CRITICAL: Update the main request object so that middlewares can see
	// the fully formed request (with URL, context, etc.).
	r.Request = req
	if originalGetBody != nil {
		r.Request.GetBody = originalGetBody
	}

	// Create a Response placeholder that will be filled by middleware
	var response *Response

	// Prepare the middleware chain
	middlewares := append(r.client.middlewares, r.middlewares...)
	// The base handler is the actual HTTP call. Middlewares wrap around this handler.
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

	// Apply middlewares in reverse order to create the chain:
	// Handler -> MiddlewareN -> ... -> Middleware1
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	// Execute the middleware chain with our Request object
	response, err = handler(r)

	// Handle errors
	if err != nil {
		return nil, err
	}

	// Set response to request
	r.SetResponse(response)

	return response, nil
}

// Do executes the request.
//
// Deprecated: This method is deprecated and will be removed in a future version.
// Please use the HTTP method-specific functions like Get, Post, etc., instead.
// For example, instead of `req.Method("GET").Do()`, use `req.Get(url)`.
func (r *Request) Do() (*Response, error) {
	if r.Request == nil || r.Request.URL == nil {
		return nil, merror.New("mclient: request URL is not set")
	}
	return r.Send(r.Request.URL.String())
}
