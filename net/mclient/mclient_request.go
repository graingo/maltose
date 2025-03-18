package mclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Request wraps the http.Request with additional features.
type Request struct {
	*http.Request
	client      *Client
	middlewares []MiddlewareFunc
	queryParams url.Values
	formParams  url.Values
}

// NewRequest creates and returns a new request object.
func (c *Client) NewRequest() *Request {
	return &Request{
		client:      c,
		middlewares: make([]MiddlewareFunc, 0),
		queryParams: make(url.Values),
		formParams:  make(url.Values),
	}
}

// Get sends a GET request and returns the response object.
func (c *Client) Get(ctx context.Context, url string) (*Response, error) {
	return c.NewRequest().Get(ctx, url)
}

// Post sends a POST request and returns the response object.
func (c *Client) Post(ctx context.Context, url string, data interface{}) (*Response, error) {
	return c.NewRequest().Post(ctx, url, data)
}

// Put sends a PUT request and returns the response object.
func (c *Client) Put(ctx context.Context, url string, data interface{}) (*Response, error) {
	return c.NewRequest().Put(ctx, url, data)
}

// Delete sends a DELETE request and returns the response object.
func (c *Client) Delete(ctx context.Context, url string) (*Response, error) {
	return c.NewRequest().Delete(ctx, url)
}

// Get sends GET request and returns the response object.
func (r *Request) Get(ctx context.Context, url string) (*Response, error) {
	return r.DoRequest(ctx, http.MethodGet, url)
}

// Post sends POST request and returns the response object.
func (r *Request) Post(ctx context.Context, url string, data interface{}) (*Response, error) {
	return r.ContentType("application/json").Data(data).DoRequest(ctx, http.MethodPost, url)
}

// Put sends PUT request and returns the response object.
func (r *Request) Put(ctx context.Context, url string, data interface{}) (*Response, error) {
	return r.ContentType("application/json").Data(data).DoRequest(ctx, http.MethodPut, url)
}

// Delete sends DELETE request and returns the response object.
func (r *Request) Delete(ctx context.Context, url string) (*Response, error) {
	return r.DoRequest(ctx, http.MethodDelete, url)
}

// Header sets an HTTP header for the request.
func (r *Request) Header(key, value string) *Request {
	if r.Request == nil {
		r.Request = &http.Request{
			Header: make(http.Header),
		}
	}
	r.Request.Header.Set(key, value)
	return r
}

// ContentType sets the Content-Type header for the request.
func (r *Request) ContentType(contentType string) *Request {
	return r.Header("Content-Type", contentType)
}

// Data sets the request data.
func (r *Request) Data(data interface{}) *Request {
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
			// 记录错误但继续执行
			// 未来可以添加错误日志或使用相关的包来记录错误
			return r
		}
		r.Request.Body = io.NopCloser(bytes.NewReader(jsonBytes))
		if r.Request.Header.Get("Content-Type") == "" {
			r.ContentType("application/json")
		}
	}
	return r
}

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

// DoRequest sends the request and returns the response.
func (r *Request) DoRequest(ctx context.Context, method string, urlPath string) (*Response, error) {
	var (
		req  *http.Request
		resp *http.Response
		err  error
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

	// 处理查询参数
	if len(r.queryParams) > 0 {
		if strings.Contains(fullURL, "?") {
			fullURL = fullURL + "&" + r.queryParams.Encode()
		} else {
			fullURL = fullURL + "?" + r.queryParams.Encode()
		}
	}

	// 处理表单参数
	var body io.Reader
	if len(r.formParams) > 0 {
		// 优先使用表单数据
		body = strings.NewReader(r.formParams.Encode())
		if r.Request == nil {
			r.Request = &http.Request{
				Header: make(http.Header),
			}
		}
		r.ContentType("application/x-www-form-urlencoded")
	} else if r.Request != nil && r.Request.Body != nil {
		// 其次使用已设置的请求体
		body = r.Request.Body
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

	// Prepare the middleware chain
	middlewares := append(r.client.middlewares, r.middlewares...)
	if len(middlewares) > 0 {
		handler := func(req *http.Request) (*http.Response, error) {
			return r.client.Do(req)
		}

		// Apply middlewares in reverse order
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}

		// Execute the middleware chain
		resp, err = handler(req)
	} else {
		// Direct request without middleware
		resp, err = r.client.Do(req)
	}

	// Handle errors
	if err != nil {
		return nil, err
	}

	// Return the response
	return &Response{
		Response:      resp,
		request:       req,
		clientRequest: r,
	}, nil
}
