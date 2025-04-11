package mclient

import (
	"fmt"
	"log"
	"time"

	"github.com/graingo/maltose/internal/intlog"
	"github.com/graingo/maltose/os/mlog"
)

// ExampleMiddlewareUsage demonstrates how to use middleware
func ExampleMiddlewareUsage() {
	// Create client
	client := New()

	// Add custom logging middleware
	client.Use(func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			fmt.Printf("Request: %s %s\n", req.Request.Method, req.Request.URL.String())

			// Execute the request
			start := time.Now()
			resp, err := next(req)
			duration := time.Since(start)

			// Log after request completion
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Response: %d, Duration: %v\n", resp.StatusCode, duration)
			}

			return resp, err
		}
	})

	// Send request with middleware applied
	resp, err := client.R().GET("https://api.example.com/users")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// Process response
	fmt.Printf("Got %d bytes of data\n", len(resp.ReadAll()))
}

// ExampleUsingLogMiddleware demonstrates how to use the log middleware
func ExampleUsingLogMiddleware() {
	// Create client
	client := New()

	// Add log middleware
	logger := mlog.DefaultLogger()
	client.Use(MiddlewareLog(logger))

	// Send request
	resp, err := client.R().GET("https://api.example.com/status")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Got status: %d\n", resp.StatusCode)
}

// ExampleUsingRateLimit demonstrates how to use the rate limit middleware
func ExampleUsingRateLimit() {
	// Create client
	client := New()

	// Add rate limiting middleware - limit to 5 requests per second
	client.Use(MiddlewareRateLimit(RateLimitConfig{
		RequestsPerSecond: 5,
		Burst:             2,
	}))

	// Send multiple requests - they will be rate limited
	for i := 0; i < 10; i++ {
		go func(index int) {
			resp, err := client.R().GET("https://api.example.com/users")
			if err != nil {
				intlog.Printf(nil, "Request %d failed: %v", index, err)
				return
			}
			defer resp.Close()

			intlog.Printf(nil, "Request %d completed with status: %d", index, resp.StatusCode)
		}(i)
	}

	// Allow time for requests to complete
	time.Sleep(3 * time.Second)
}

// ExampleCustomMiddleware demonstrates how to create custom middleware
func ExampleCustomMiddleware() {
	// Create custom authentication middleware
	authMiddleware := func(token string) MiddlewareFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(req *Request) (*Response, error) {
				// Add authentication header to every request
				req.SetHeader("Authorization", "Bearer "+token)
				return next(req)
			}
		}
	}

	// Create retry with backoff middleware
	retryWithBackoff := func(maxRetries int) MiddlewareFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(req *Request) (*Response, error) {
				var resp *Response
				var err error

				// Try request with increasing backoff
				for attempt := 0; attempt <= maxRetries; attempt++ {
					if attempt > 0 {
						// Exponential backoff
						backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
						time.Sleep(backoff)
						intlog.Printf(nil, "Retrying request (attempt %d/%d) after %v",
							attempt, maxRetries, backoff)
					}

					resp, err = next(req)

					// Success or non-retriable error
					if err == nil && resp.StatusCode < 500 {
						return resp, err
					}

					// Close the response if we got one but will retry
					if resp != nil {
						resp.Close()
					}
				}

				return resp, err
			}
		}
	}

	// Create client with custom middlewares
	client := New()
	client.Use(
		authMiddleware("user-token-123"),
		retryWithBackoff(3),
	)

	// Send request
	resp, err := client.R().GET("https://api.example.com/secure-resource")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Secure request completed with status: %d\n", resp.StatusCode)
}

// min 辅助函数，返回两个整数中较小的一个
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
