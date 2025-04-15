package mclient

import (
	"math/rand"
	"net/http"
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

// Do performs the HTTP request using the underlying HTTP client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
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

// calculateRetryDelay calculates the delay for the next retry attempt.
func (r *Request) calculateRetryDelay(attempt int) time.Duration {
	// If no retry config, use simple interval
	if r.retryConfig == (RetryConfig{}) {
		return r.retryInterval
	}

	// Calculate exponential backoff
	delay := r.retryConfig.BaseInterval
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * r.retryConfig.BackoffFactor)
		if delay > r.retryConfig.MaxInterval {
			delay = r.retryConfig.MaxInterval
			break
		}
	}

	// Add jitter
	if r.retryConfig.JitterFactor > 0 {
		jitter := time.Duration(float64(delay) * r.retryConfig.JitterFactor * (rand.Float64()*2 - 1))
		delay += jitter
		if delay < 0 {
			delay = 0
		}
	}

	return delay
}

// Do executes the request with retry logic.
func (r *Request) Do() (*Response, error) {
	var resp *Response
	var err error

	// Initialize random seed if needed
	if r.retryConfig != (RetryConfig{}) && r.retryConfig.JitterFactor > 0 {
		rand.Seed(time.Now().UnixNano())
	}

	// Try the request up to retryCount + 1 times
	for attempt := 0; attempt <= r.retryCount; attempt++ {
		if attempt > 0 {
			// Calculate delay for this attempt
			delay := r.calculateRetryDelay(attempt)
			time.Sleep(delay)
		}

		// Execute the request
		resp, err = r.DoRequest(r.ctx, r.Request.Method, r.Request.URL.String())

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
