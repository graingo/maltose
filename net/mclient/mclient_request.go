package mclient

import (
	"net/http"
)

// Request 请求信息
type Request struct {
	Method  string
	URL     string
	Headers http.Header
	Body    []byte
}
