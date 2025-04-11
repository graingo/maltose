package mclient

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleCustomConfig demonstrates how to use advanced configuration options
func ExampleCustomConfig() {
	// Create client with custom configuration
	client := NewWithConfig(ClientConfig{
		Timeout: 10 * time.Second,
		BaseURL: "https://api.example.com",
		Header: http.Header{
			"User-Agent": []string{"MaltoseClient/1.0"},
			"Accept":     []string{"application/json"},
		},
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send request with context
	resp, err := client.R().
		SetContext(ctx).
		GET("/users")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Request successful, status code: %d\n", resp.StatusCode)
}

// ExampleClientConfiguration demonstrates various client configuration options
func ExampleClientConfiguration() {
	// Create client with default configuration
	client := New()

	// Set base URL
	client.SetConfig(ClientConfig{
		BaseURL: "https://api.example.com/v1",
	})

	// Set custom transport
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}
	client.SetTransport(transport)

	// Send request, base URL will be prepended automatically
	resp, err := client.R().GET("/users")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Request to %s successful\n", "https://api.example.com/v1/users")
}

// ExampleRequestConfiguration demonstrates request-level configuration
func ExampleRequestConfiguration() {
	client := New()

	// Set request timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Configure retry policy
	resp, err := client.R().
		SetContext(ctx).
		SetRetry(3, 1*time.Second).
		SetRetryCondition(func(resp *http.Response, err error) bool {
			// Retry on network errors
			if err != nil {
				return true
			}
			// Retry on server errors
			if resp != nil && resp.StatusCode >= 500 {
				return true
			}
			return false
		}).
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed after retries: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Request completed with status: %d\n", resp.StatusCode)
}

// ExampleCloning demonstrates how to clone a client and modify its configuration
func ExampleCloning() {
	// Create base client
	baseClient := New()
	baseClient.SetConfig(ClientConfig{
		BaseURL: "https://api.example.com",
		Timeout: 10 * time.Second,
	})

	// Clone client and modify configuration
	adminClient := baseClient.Clone()
	adminClient.SetConfig(ClientConfig{
		BaseURL: "https://admin.example.com",
		Header: http.Header{
			"Authorization": []string{"Bearer admin-token"},
		},
	})

	// Use base client
	resp1, err := baseClient.R().GET("/users")
	if err != nil {
		log.Printf("Base client request failed: %v", err)
	} else {
		defer resp1.Close()
		fmt.Printf("Base client status: %d\n", resp1.StatusCode)
	}

	// Use admin client
	resp2, err := adminClient.R().GET("/dashboard")
	if err != nil {
		log.Printf("Admin client request failed: %v", err)
	} else {
		defer resp2.Close()
		fmt.Printf("Admin client status: %d\n", resp2.StatusCode)
	}
}
