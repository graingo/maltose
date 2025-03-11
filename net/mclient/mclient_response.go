package mclient

import (
	"net/http"
)

// Response is the struct for client request response.
type Response struct {
	*http.Response                   // Response is the underlying http.Response object of certain request.
	request        *http.Request     // Request is the underlying http.Request object of certain request.
	requestBody    []byte            // The body bytes of certain request, only available in Dump feature.
	cookies        map[string]string // Response cookies, which are only parsed once.
}
