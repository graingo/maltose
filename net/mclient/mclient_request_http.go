package mclient

import (
	"net/http"
)

// Get sets the method to GET and executes the request.
func (r *Request) Get(url string) (*Response, error) {
	return r.Method(http.MethodGet).Send(url)
}

// Post sets the method to POST and executes the request.
func (r *Request) Post(url string) (*Response, error) {
	return r.Method(http.MethodPost).Send(url)
}

// Put sets the method to PUT and executes the request.
func (r *Request) Put(url string) (*Response, error) {
	return r.Method(http.MethodPut).Send(url)
}

// Delete sets the method to DELETE and executes the request.
func (r *Request) Delete(url string) (*Response, error) {
	return r.Method(http.MethodDelete).Send(url)
}

// Patch sets the method to PATCH and executes the request.
func (r *Request) Patch(url string) (*Response, error) {
	return r.Method(http.MethodPatch).Send(url)
}

// Head sets the method to HEAD and executes the request.
func (r *Request) Head(url string) (*Response, error) {
	return r.Method(http.MethodHead).Send(url)
}

// Options sets the method to OPTIONS and executes the request.
func (r *Request) Options(url string) (*Response, error) {
	return r.Method(http.MethodOptions).Send(url)
}
