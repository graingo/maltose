package mclient

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleRetryMechanism demonstrates the basic retry mechanism.
func ExampleRetryMechanism() {
	client := New()

	// Use simple retry configuration
	resp, err := client.R().
		SetRetrySimple(3, time.Second).
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed after retries: %v", err)
		return
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode)
}

// ExampleCustomRetryCondition demonstrates how to use custom retry conditions.
func ExampleCustomRetryCondition() {
	client := New()

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
	config := RetryConfig{
		Count:         3,                // Maximum 3 retries
		BaseInterval:  time.Second,      // Base interval 1 second
		MaxInterval:   30 * time.Second, // Maximum interval 30 seconds
		BackoffFactor: 2.0,              // Exponential backoff factor 2.0
		JitterFactor:  0.1,              // Random jitter factor 0.1
	}

	// Send request
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

// ExampleChainedRequests demonstrates how to chain multiple requests with retry.
func ExampleChainedRequests() {
	client := New()

	// Configure retry strategy
	config := RetryConfig{
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
