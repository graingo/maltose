package mclient_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/graingo/maltose/net/mclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCustomMiddleware tests custom middleware functionality
func TestCustomMiddleware(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert custom header added by middleware
		assert.Equal(t, "middleware-value", r.Header.Get("X-Custom-Header"), "Expected X-Custom-Header: middleware-value")

		// Write response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer server.Close()

	// Create client
	client := mclient.New()

	// Define custom middleware that adds a header to requests
	headerMiddleware := func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(req *mclient.Request) (*mclient.Response, error) {
			// Add custom header to request
			req.SetHeader("X-Custom-Header", "middleware-value")

			// Call the next handler
			return next(req)
		}
	}

	// Add middleware to client
	client.Use(headerMiddleware)

	// Send request
	resp, err := client.R().GET(server.URL)

	// Assert response
	require.NoError(t, err, "Should not return error")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")
}

// TestMultipleMiddlewares tests that multiple middlewares are executed in the right order
func TestMultipleMiddlewares(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert headers from middlewares
		assert.Equal(t, "first", r.Header.Get("X-Order-1"), "Expected X-Order-1: first")
		assert.Equal(t, "second", r.Header.Get("X-Order-2"), "Expected X-Order-2: second")

		// Write response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer server.Close()

	// Create client
	client := mclient.New()

	// Track middleware execution order
	executionOrder := []string{}

	// First middleware
	middleware1 := func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(req *mclient.Request) (*mclient.Response, error) {
			executionOrder = append(executionOrder, "middleware1")
			req.SetHeader("X-Order-1", "first")
			resp, err := next(req)
			executionOrder = append(executionOrder, "middleware1-after")
			return resp, err
		}
	}

	// Second middleware
	middleware2 := func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(req *mclient.Request) (*mclient.Response, error) {
			executionOrder = append(executionOrder, "middleware2")
			req.SetHeader("X-Order-2", "second")
			resp, err := next(req)
			executionOrder = append(executionOrder, "middleware2-after")
			return resp, err
		}
	}

	// Add middlewares to client
	client.Use(middleware1, middleware2)

	// Send request
	resp, err := client.R().GET(server.URL)

	// Assert response
	require.NoError(t, err, "Should not return error")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Assert middleware execution order (should follow onion model)
	expectedOrder := []string{
		"middleware1",
		"middleware2",
		"middleware2-after",
		"middleware1-after",
	}

	// Compare execution order
	assert.Equal(t, len(expectedOrder), len(executionOrder), "Expected %d middleware executions", len(expectedOrder))
	for i, step := range expectedOrder {
		if i < len(executionOrder) {
			assert.Equal(t, step, executionOrder[i], "Middleware execution order incorrect at step %d", i)
		}
	}
}

// TestLogMiddleware tests the logging middleware
func TestLogMiddleware(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`OK`))
	}))
	defer server.Close()

	// Create buffer to capture logs
	var logBuffer bytes.Buffer

	// Mock logger middleware which writes to our buffer
	logMiddleware := func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(req *mclient.Request) (*mclient.Response, error) {
			// Log request to our buffer
			urlStr := "<no url>"
			if req.Request != nil && req.Request.URL != nil {
				urlStr = req.Request.URL.String()
			}
			fmt.Fprintf(&logBuffer, "Request: %s %s\n", req.Request.Method, urlStr)

			// Execute request
			resp, err := next(req)

			// Log response to our buffer
			if err != nil {
				fmt.Fprintf(&logBuffer, "Error: %v\n", err)
			} else {
				fmt.Fprintf(&logBuffer, "Response: %d\n", resp.StatusCode)
			}

			return resp, err
		}
	}

	// Create client with our custom logging middleware
	client := mclient.New()
	client.Use(logMiddleware)

	// Send request
	_, err := client.R().GET(server.URL)
	require.NoError(t, err, "Should not return error")

	// Check that logs were captured
	logOutput := logBuffer.String()

	// Check for request log
	assert.Contains(t, logOutput, "Request:", "Expected request to be logged")

	// Check for response log
	assert.Contains(t, logOutput, "Response:", "Expected response to be logged")
}

// TestRateLimitMiddleware tests the rate limiting middleware
func TestRateLimitMiddleware(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with rate limiting middleware
	client := mclient.New()

	// Add rate limiting middleware with very low limits
	client.Use(mclient.MiddlewareRateLimit(mclient.RateLimitConfig{
		RequestsPerSecond: 2, // Allow only 2 requests per second
		Burst:             1, // Allow only 1 burst request
	}))

	// Track request times
	var requestTimes []time.Time

	// Make 5 requests
	for i := 0; i < 5; i++ {
		requestTimes = append(requestTimes, time.Now())
		_, _ = client.R().GET(server.URL)
	}

	// Check time differences between first and last request
	// Should be rate limited to ~2 per second
	totalDuration := requestTimes[len(requestTimes)-1].Sub(requestTimes[0])

	// With limit of 2/sec, 5 requests should take at least 2 seconds
	minimumExpectedDuration := 2 * time.Second

	assert.GreaterOrEqual(t, totalDuration, minimumExpectedDuration,
		"Rate limiting not working properly. All requests completed in %v, expected at least %v",
		totalDuration, minimumExpectedDuration)
}
