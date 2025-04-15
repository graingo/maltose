package mclient

import (
	"net/http"
)

// GET sets the method to GET and executes the request.
func (r *Request) GET(url string) (*Response, error) {
	return r.Method(http.MethodGet).Send(url)
}

// POST sets the method to POST and executes the request.
func (r *Request) POST(url string) (*Response, error) {
	return r.Method(http.MethodPost).Send(url)
}

// PUT sets the method to PUT and executes the request.
func (r *Request) PUT(url string) (*Response, error) {
	return r.Method(http.MethodPut).Send(url)
}

// DELETE sets the method to DELETE and executes the request.
func (r *Request) DELETE(url string) (*Response, error) {
	return r.Method(http.MethodDelete).Send(url)
}

// PATCH sets the method to PATCH and executes the request.
func (r *Request) PATCH(url string) (*Response, error) {
	return r.Method(http.MethodPatch).Send(url)
}

// HEAD sets the method to HEAD and executes the request.
func (r *Request) HEAD(url string) (*Response, error) {
	return r.Method(http.MethodHead).Send(url)
}

// OPTIONS sets the method to OPTIONS and executes the request.
func (r *Request) OPTIONS(url string) (*Response, error) {
	return r.Method(http.MethodOptions).Send(url)
}
