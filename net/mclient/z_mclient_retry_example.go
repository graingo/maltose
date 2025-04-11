package mclient

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleRetryMechanism demonstrates how to use the retry mechanism
func ExampleRetryMechanism() {
	// Create client
	client := New()

	// Send request with retry configuration
	resp, err := client.R().
		SetRetry(3, time.Second). // 3 retries with 1 second interval
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed after retries: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Request successful after retries, status: %d\n", resp.StatusCode)
}

// ExampleCustomRetryCondition demonstrates how to use custom retry conditions
func ExampleCustomRetryCondition() {
	// Create client
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

		// Don't retry for other status codes
		return false
	}

	// Send request with custom retry configuration
	resp, err := client.R().
		SetRetry(5, 500*time.Millisecond).
		SetRetryCondition(customRetryCondition).
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed after custom retries: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Request successful with custom retry, status: %d\n", resp.StatusCode)
}

// ExampleChainedRequests demonstrates how to chain multiple requests
func ExampleChainedRequests() {
	// Create client with common settings
	client := New()

	// Configure client with reasonable retry settings
	client.SetConfig(ClientConfig{
		Timeout: 30 * time.Second,
		Header: http.Header{
			"Accept":     []string{"application/json"},
			"User-Agent": []string{"MClient/1.0"},
		},
	})

	// First request - get a list of items
	type Item struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	var items []Item

	resp1, err := client.R().
		SetResult(&items).
		SetRetry(3, time.Second).
		GET("https://api.example.com/items")

	if err != nil {
		log.Printf("Failed to fetch items: %v", err)
		return
	}
	defer resp1.Close()

	if !resp1.IsSuccess() {
		log.Printf("Failed to fetch items, status: %d", resp1.StatusCode)
		return
	}

	fmt.Printf("Found %d items\n", len(items))

	// If we have items, make a second request for details of the first item
	if len(items) > 0 {
		type ItemDetails struct {
			ID          int      `json:"id"`
			Name        string   `json:"name"`
			Price       float64  `json:"price"`
			Description string   `json:"description"`
			Categories  []string `json:"categories"`
			InStock     bool     `json:"in_stock"`
		}
		var details ItemDetails

		resp2, err := client.R().
			SetResult(&details).
			SetRetry(3, time.Second).
			GET(fmt.Sprintf("https://api.example.com/items/%d", items[0].ID))

		if err != nil {
			log.Printf("Failed to fetch item details: %v", err)
			return
		}
		defer resp2.Close()

		if resp2.IsSuccess() {
			fmt.Printf("Item details: %s - $%.2f\n", details.Name, details.Price)
			fmt.Printf("Description: %s\n", details.Description)
			fmt.Printf("Categories: %v\n", details.Categories)
			fmt.Printf("In stock: %v\n", details.InStock)
		} else {
			log.Printf("Failed to fetch item details, status: %d", resp2.StatusCode)
		}
	}
}
