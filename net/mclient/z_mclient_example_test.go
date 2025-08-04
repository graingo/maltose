package mclient_test

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/graingo/maltose/net/mclient"
)

// Example demonstrates basic usage of the client.
// This is a runnable example, but it will not make a real network call
// as the URL is a placeholder. To make it a "godoc" example,
// we'd typically mock the server.
func Example() {
	client := mclient.New()

	// In a real scenario, you would handle the response and error.
	_, err := client.R().
		SetHeader("Accept", "application/json").
		Get("https://api.example.com/users")

	if err != nil {
		// This will likely print an error in a test environment, which is expected.
		fmt.Println("Request intended to fail for example purposes.")
	}
	// Output:
	// Request intended to fail for example purposes.
}

// Example_jSON demonstrates JSON request and response handling.
func Example_jSON() {
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

	// Send request (this will fail as the URL is not real)
	_, err := client.R().
		SetBody(user).
		SetResult(&result).
		Post("https://api.example.com/users")

	if err != nil {
		fmt.Println("JSON request example finished.")
	}

	// In a real case, you might print the result:
	// fmt.Printf("Created user: %s (ID: %d)\n", result.Name, result.ID)

	// Output:
	// JSON request example finished.
}

// Example_retry demonstrates retry mechanism.
func Example_retry() {
	client := mclient.New()

	// Configure retry strategy
	config := mclient.RetryConfig{
		Count:         3,
		BaseInterval:  time.Second,
		MaxInterval:   30 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
	}

	// Send request with retry (this will fail as the URL is not real)
	_, err := client.R().
		SetRetry(config).
		Get("https://api.example.com/users")

	if err != nil {
		fmt.Println("Retry example finished.")
	}

	// Output:
	// Retry example finished.
}

// Example_customRetryCondition demonstrates custom retry conditions.
func Example_customRetryCondition() {
	client := mclient.New()

	// Define custom retry condition
	customRetryCondition := func(resp *http.Response, err error) bool {
		// Retry on network errors
		if err != nil {
			return true
		}
		// Retry on server errors (5xx) or rate limiting (429)
		if resp != nil && (resp.StatusCode >= 500 || resp.StatusCode == 429) {
			return true
		}
		return false
	}

	config := mclient.RetryConfig{Count: 3}

	_, err := client.R().
		SetRetry(config).
		SetRetryCondition(customRetryCondition).
		Get("https://api.example.com/users")

	if err != nil {
		fmt.Println("Custom retry example finished.")
	}
	// Output:
	// Custom retry example finished.
}

// Example_middleware demonstrates middleware usage.
func Example_middleware() {
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
			}
			return resp, err
		}
	}))

	_, err := client.R().Get("https://api.example.com/users")

	if err != nil {
		fmt.Println("Middleware example finished.")
	}
	// Output:
	// Middleware example finished.
}
