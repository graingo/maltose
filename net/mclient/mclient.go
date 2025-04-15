package mclient

import (
	"net/http"
	"net/url"
	"time"
)

// Client is an HTTP client with enhanced features.
type Client struct {
	client      *http.Client     // HTTP client for the request.
	config      ClientConfig     // Default configuration for the client.
	middlewares []MiddlewareFunc // Middleware functions.
}

// New creates and returns a new HTTP client object.
func New() *Client {
	c := &Client{
		client: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   30 * time.Second,
		},
		middlewares: make([]MiddlewareFunc, 0),
	}

	// Add default internal middlewares
	c.Use(
		internalMiddlewareRecovery(),
		internalMiddlewareTrace(),
		internalMiddlewareMetric(),
	)

	return c
}

// NewWithConfig creates and returns a client with given config.
func NewWithConfig(config ClientConfig) *Client {
	c := New()
	c.config = config

	// Apply configuration to http.Client
	if config.Timeout > 0 {
		c.client.Timeout = config.Timeout
	}
	if config.Transport != nil {
		c.client.Transport = config.Transport
	}

	return c
}

// Use adds middleware handlers to the client.
func (c *Client) Use(middlewares ...MiddlewareFunc) *Client {
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// Clone creates and returns a copy of the current client.
func (c *Client) Clone() *Client {
	newClient := New()
	newClient.client = &http.Client{
		Transport: c.client.Transport,
		Timeout:   c.client.Timeout,
	}
	newClient.config = c.config
	newClient.middlewares = append(newClient.middlewares, c.middlewares...)
	return newClient
}

// do performs the HTTP request using the underlying HTTP client.
// This is an internal method used by the middleware chain.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying the original request
	reqCopy := req.Clone(req.Context())

	// Apply client configuration
	if c.config.Header != nil && reqCopy.Header == nil {
		reqCopy.Header = make(http.Header)
	}

	for k, v := range c.config.Header {
		if reqCopy.Header.Get(k) == "" && len(v) > 0 {
			reqCopy.Header.Set(k, v[0])
		}
	}

	// Execute request
	return c.client.Do(reqCopy)
}

// GetClient returns the underlying http.Client.
func (c *Client) GetClient() *http.Client {
	return c.client
}

// SetTransport sets the client transport.
func (c *Client) SetTransport(transport http.RoundTripper) *Client {
	c.client.Transport = transport
	c.config.Transport = transport
	return c
}

// SetConfig sets the client configuration.
func (c *Client) SetConfig(config ClientConfig) *Client {
	c.config = config

	// Apply configuration to HTTP client
	if config.Timeout > 0 {
		c.client.Timeout = config.Timeout
	}
	if config.Transport != nil {
		c.client.Transport = config.Transport
	}

	return c
}

// NewRequest creates and returns a new request object.
func (c *Client) NewRequest() *Request {
	return &Request{
		client:      c,
		middlewares: make([]MiddlewareFunc, 0),
		queryParams: make(url.Values),
		formParams:  make(url.Values),
		response:    &Response{},
	}
}

// R returns a new request object bound to this client for chain calls.
func (c *Client) R() *Request {
	return c.NewRequest()
}
