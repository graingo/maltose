package mclient

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleComprehensive demonstrates a comprehensive usage example of mclient
func ExampleComprehensive() {
	// Create client with base configuration
	client := NewWithConfig(ClientConfig{
		BaseURL: "https://api.example.com/v1",
		Timeout: 30 * time.Second,
		Header: http.Header{
			"User-Agent": []string{"MaltoseClient/1.0"},
			"Accept":     []string{"application/json"},
		},
	})

	// Add logging middleware
	client.Use(func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			startTime := time.Now()

			// Log request
			fmt.Printf("[INFO] Request: %s %s\n",
				req.Request.Method, req.Request.URL.String())

			// Execute request
			resp, err := next(req)

			// Log response or error
			duration := time.Since(startTime)
			if err != nil {
				fmt.Printf("[ERROR] Request failed: %v (took %v)\n",
					err, duration)
			} else {
				fmt.Printf("[INFO] Response: Status=%d (took %v)\n",
					resp.StatusCode, duration)
			}

			return resp, err
		}
	})

	// Add custom authentication middleware
	client.Use(func(next HandlerFunc) HandlerFunc {
		return func(req *Request) (*Response, error) {
			// Add auth header to all requests
			req.SetHeader("Authorization", "Bearer your-access-token")
			return next(req)
		}
	})

	// Add rate limiting middleware
	client.Use(MiddlewareRateLimit(RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             5,
	}))

	// Create request context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Define response structures
	type User struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}

	type ErrorResponse struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}

	// Prepare result and error containers
	var user User
	var errorResponse ErrorResponse

	// Send GET request with all configurations
	resp, err := client.R().
		SetContext(ctx).
		SetResult(&user).
		SetError(&errorResponse).
		SetRetry(3, 500*time.Millisecond).
		SetRetryCondition(func(resp *http.Response, err error) bool {
			// Retry on network errors or server errors
			return err != nil || (resp != nil && resp.StatusCode >= 500)
		}).
		SetQuery("include", "profile").
		SetQuery("fields", "id,name,email,created_at").
		GET("/users/123")

	// Handle errors first
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// Handle non-success status codes
	if !resp.IsSuccess() {
		log.Printf("API error: %s (code: %d)",
			errorResponse.Message, errorResponse.Code)
		return
	}

	// Process successful response
	fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
	fmt.Printf("Created: %s\n", user.CreatedAt)
}
