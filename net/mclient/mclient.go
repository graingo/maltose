package mclient

import (
	"net/http"
	"time"
)

// Client is an HTTP client with enhanced features.
type Client struct {
	// HTTP client for the request.
	client *http.Client
	// Default configuration for the client.
	config ClientConfig
	// Middleware functions.
	middlewares []MiddlewareFunc
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
		internalMiddlewareClientTrace(),
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

// Do performs the HTTP request using the underlying HTTP client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// 复制请求以避免修改原始请求
	reqCopy := req.Clone(req.Context())

	// 应用客户端配置
	if c.config.Header != nil && reqCopy.Header == nil {
		reqCopy.Header = make(http.Header)
	}

	for k, v := range c.config.Header {
		if reqCopy.Header.Get(k) == "" && len(v) > 0 {
			reqCopy.Header.Set(k, v[0])
		}
	}

	// 执行请求
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

	// 应用配置到HTTP客户端
	if config.Timeout > 0 {
		c.client.Timeout = config.Timeout
	}
	if config.Transport != nil {
		c.client.Transport = config.Transport
	}

	return c
}
