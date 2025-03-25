package mclient

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

// Initialize the random number generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

// -----------------------------------------------------------------------------
// Common interceptors
// -----------------------------------------------------------------------------

// RequestID creates an interceptor that adds a request ID header.
func RequestID(headerName string) RequestInterceptor {
	if headerName == "" {
		headerName = "X-Request-ID"
	}

	return func(ctx context.Context, req *http.Request) (*http.Request, error) {
		// Only add if not already present
		if req.Header.Get(headerName) == "" {
			req.Header.Set(headerName, generateRequestID())
		}
		return req, nil
	}
}

// UserAgent creates an interceptor that sets the User-Agent header.
func UserAgent(userAgent string) RequestInterceptor {
	return func(ctx context.Context, req *http.Request) (*http.Request, error) {
		if req.Header.Get("User-Agent") == "" {
			req.Header.Set("User-Agent", userAgent)
		}
		return req, nil
	}
}

// Authentication creates an interceptor that adds authentication headers.
func Authentication(authType string, getToken func() string) RequestInterceptor {
	return func(ctx context.Context, req *http.Request) (*http.Request, error) {
		token := getToken()
		if token != "" {
			req.Header.Set("Authorization", authType+" "+token)
		}
		return req, nil
	}
}

// Helpers
func generateRequestID() string {
	// In a real implementation, you might want to use a UUID library
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
