package mclient_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/graingo/maltose/net/mclient"
)

// Example demonstrates basic usage of the client
func Example() {
	client := mclient.New()

	// Send a simple GET request
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode)
}

// ExampleJSON demonstrates JSON request and response handling
func ExampleJSON() {
	client := mclient.New()

	// Define request and response structures
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	type CreateResponse struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	// Prepare request data
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Prepare result container
	var result CreateResponse

	// Send request
	_, err := client.R().
		SetBody(user).
		SetResult(&result).
		POST("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	fmt.Printf("Created user: %s (ID: %d)\n", result.Name, result.ID)
}

// ExampleRetry demonstrates retry mechanism
func ExampleRetry() {
	client := mclient.New()

	// Configure retry strategy
	config := mclient.RetryConfig{
		Count:         3,                // Maximum 3 retries
		BaseInterval:  time.Second,      // Base interval 1 second
		MaxInterval:   30 * time.Second, // Maximum interval 30 seconds
		BackoffFactor: 2.0,              // Exponential backoff factor 2.0
		JitterFactor:  0.1,              // Random jitter factor 0.1
	}

	// Send request with retry
	resp, err := client.R().
		SetRetry(config).
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed after retries: %v", err)
		return
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode)
}

// ExampleCustomRetryCondition demonstrates custom retry conditions
func ExampleCustomRetryCondition() {
	client := mclient.New()

	// Define custom retry condition
	customRetryCondition := func(resp *http.Response, err error) bool {
		// Retry on network errors
		if err != nil {
			return true
		}

		// Retry on server errors (5xx)
		if resp != nil && resp.StatusCode >= 500 {
			return true
		}

		// Retry on rate limiting (429)
		if resp != nil && resp.StatusCode == 429 {
			return true
		}

		return false
	}

	// Configure retry strategy
	config := mclient.RetryConfig{
		Count:         3,                // Maximum 3 retries
		BaseInterval:  time.Second,      // Base interval 1 second
		MaxInterval:   30 * time.Second, // Maximum interval 30 seconds
		BackoffFactor: 2.0,              // Exponential backoff factor 2.0
		JitterFactor:  0.1,              // Random jitter factor 0.1
	}

	// Send request with custom retry condition
	resp, err := client.R().
		SetRetry(config).
		SetRetryCondition(customRetryCondition).
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed after retries: %v", err)
		return
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode)
}

// ExampleMiddleware demonstrates middleware usage
func ExampleMiddleware() {
	client := mclient.New()

	// Add auth middleware
	client.Use(mclient.MiddlewareFunc(func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(req *mclient.Request) (*mclient.Response, error) {
			req.SetHeader("Authorization", "Bearer test-token")
			return next(req)
		}
	}))

	// Add logging middleware
	client.Use(mclient.MiddlewareFunc(func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(req *mclient.Request) (*mclient.Response, error) {
			log.Printf("Sending request to %s", req.Request.URL.String())
			resp, err := next(req)
			if err != nil {
				log.Printf("Request failed: %v", err)
			} else {
				log.Printf("Response status: %d", resp.StatusCode)
			}
			return resp, err
		}
	}))

	// Send request
	resp, err := client.R().GET("https://api.example.com/users")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode)
}

// ExampleRateLimit demonstrates rate limiting middleware
func ExampleRateLimit() {
	client := mclient.New()

	// Add rate limit middleware (2 requests per second)
	client.Use(mclient.MiddlewareRateLimit(mclient.RateLimitConfig{
		RequestsPerSecond: 2,
		Burst:             1,
	}))

	// Send multiple requests
	for i := 0; i < 3; i++ {
		resp, err := client.R().GET("https://api.example.com/users")
		if err != nil {
			log.Printf("Request %d failed: %v", i+1, err)
			continue
		}
		fmt.Printf("Request %d status: %d\n", i+1, resp.StatusCode)
	}
}

// ExampleChainedRequests demonstrates chaining multiple requests
func ExampleChainedRequests() {
	client := mclient.New()

	// Configure retry strategy
	config := mclient.RetryConfig{
		Count:         3,                // Maximum 3 retries
		BaseInterval:  time.Second,      // Base interval 1 second
		MaxInterval:   30 * time.Second, // Maximum interval 30 seconds
		BackoffFactor: 2.0,              // Exponential backoff factor 2.0
		JitterFactor:  0.1,              // Random jitter factor 0.1
	}

	// Define response structures
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type UserList struct {
		Users []User `json:"users"`
	}

	// Get user list
	var userList UserList
	_, err := client.R().
		SetRetry(config).
		SetResult(&userList).
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Failed to get user list: %v", err)
		return
	}

	// Process each user's details
	for _, user := range userList.Users {
		var userDetail User
		_, err := client.R().
			SetRetry(config).
			SetResult(&userDetail).
			GET(fmt.Sprintf("https://api.example.com/users/%d", user.ID))

		if err != nil {
			log.Printf("Failed to get user %d details: %v", user.ID, err)
			continue
		}

		fmt.Printf("User %d: %s\n", userDetail.ID, userDetail.Name)
	}
}
