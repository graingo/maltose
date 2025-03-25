package mclient

import (
	"context"
	"net/http"
)

// RequestInterceptor intercepts and potentially modifies an outgoing request.
type RequestInterceptor func(context.Context, *http.Request) (*http.Request, error)

// ResponseInterceptor intercepts and potentially modifies an incoming response.
type ResponseInterceptor func(context.Context, *http.Response) (*http.Response, error)

// Interceptors holds request and response interceptors.
type Interceptors struct {
	RequestInterceptors  []RequestInterceptor
	ResponseInterceptors []ResponseInterceptor
}

// InterceptorMiddleware creates a middleware that applies the interceptors.
func InterceptorMiddleware(interceptors Interceptors) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(req *http.Request) (*http.Response, error) {
			ctx := req.Context()
			var err error

			// Apply request interceptors in order
			for _, interceptor := range interceptors.RequestInterceptors {
				if req, err = interceptor(ctx, req); err != nil {
					return nil, err
				}
			}

			// Execute the actual request
			resp, err := next(req)
			if err != nil {
				return nil, err
			}

			// Apply response interceptors in order
			for _, interceptor := range interceptors.ResponseInterceptors {
				if resp, err = interceptor(ctx, resp); err != nil {
					return nil, err
				}
			}

			return resp, nil
		}
	}
}

// -----------------------------------------------------------------------------
// Client methods for interceptors
// -----------------------------------------------------------------------------

// WithInterceptors adds request and response interceptors to the client.
func (c *Client) WithInterceptors(interceptors Interceptors) *Client {
	c.Use(InterceptorMiddleware(interceptors))
	return c
}

// OnRequest adds a request interceptor to the client.
func (c *Client) OnRequest(interceptor RequestInterceptor) *Client {
	return c.WithInterceptors(Interceptors{
		RequestInterceptors: []RequestInterceptor{interceptor},
	})
}

// OnResponse adds a response interceptor to the client.
func (c *Client) OnResponse(interceptor ResponseInterceptor) *Client {
	return c.WithInterceptors(Interceptors{
		ResponseInterceptors: []ResponseInterceptor{interceptor},
	})
}

// -----------------------------------------------------------------------------
// Request methods for interceptors
// -----------------------------------------------------------------------------

// WithInterceptors adds interceptors to the request.
func (r *Request) WithInterceptors(interceptors Interceptors) *Request {
	r.Use(InterceptorMiddleware(interceptors))
	return r
}

// OnRequest adds a request interceptor to the request.
func (r *Request) OnRequest(interceptor RequestInterceptor) *Request {
	return r.WithInterceptors(Interceptors{
		RequestInterceptors: []RequestInterceptor{interceptor},
	})
}

// OnResponse adds a response interceptor to the request.
func (r *Request) OnResponse(interceptor ResponseInterceptor) *Request {
	return r.WithInterceptors(Interceptors{
		ResponseInterceptors: []ResponseInterceptor{interceptor},
	})
}
