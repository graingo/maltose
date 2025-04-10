package mclient

// HandlerFunc defines the handler used by middleware.
type HandlerFunc func(req *Request) (*Response, error)

// MiddlewareFunc defines a function to process middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Use adds middleware handlers to the request.
func (r *Request) Use(middlewares ...MiddlewareFunc) *Request {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}
