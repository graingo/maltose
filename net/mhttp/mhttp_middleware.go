package mhttp

// MiddlewareFunc defines the middleware function type.
type MiddlewareFunc func(*Request)

// Use adds global middleware.
func (s *Server) Use(middlewares ...MiddlewareFunc) {
	s.RouterGroup.Middleware(middlewares...)
}

// Middleware alias Use.
func (s *Server) Middleware(middlewares ...MiddlewareFunc) {
	s.Use(middlewares...)
}
