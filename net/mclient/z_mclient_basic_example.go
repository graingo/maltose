package mclient

import (
	"fmt"
	"log"
)

// ExampleBasicRequest demonstrates the basic usage of mclient
func ExampleBasicRequest() {
	// Create default client
	client := New()

	// Send GET request with chain API
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// Check response status with IsSuccess
	if !resp.IsSuccess() {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
		return
	}

	// Read response content
	bodyContent := resp.ReadAllString()
	fmt.Println("Response:", bodyContent)
}

// ExampleSimpleRequests demonstrates how to send different types of HTTP requests
func ExampleSimpleRequests() {
	// Create client
	client := New()

	// Send GET request
	getResp, err := client.R().GET("https://api.example.com/users")
	if err != nil {
		log.Printf("GET request failed: %v", err)
		return
	}
	defer getResp.Close()
	fmt.Printf("GET response status: %d\n", getResp.StatusCode)

	// Send POST request
	postResp, err := client.R().
		SetBody(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		}).
		POST("https://api.example.com/users")
	if err != nil {
		log.Printf("POST request failed: %v", err)
		return
	}
	defer postResp.Close()
	fmt.Printf("POST response status: %d\n", postResp.StatusCode)

	// Send PUT request
	putResp, err := client.R().
		SetBody(map[string]interface{}{
			"name":  "John Doe Updated",
			"email": "john.updated@example.com",
		}).
		PUT("https://api.example.com/users/1")
	if err != nil {
		log.Printf("PUT request failed: %v", err)
		return
	}
	defer putResp.Close()
	fmt.Printf("PUT response status: %d\n", putResp.StatusCode)

	// Send DELETE request
	deleteResp, err := client.R().DELETE("https://api.example.com/users/1")
	if err != nil {
		log.Printf("DELETE request failed: %v", err)
		return
	}
	defer deleteResp.Close()
	fmt.Printf("DELETE response status: %d\n", deleteResp.StatusCode)
}

// ExampleResponseHandling demonstrates how to handle response content
func ExampleResponseHandling() {
	client := New()

	resp, err := client.R().GET("https://api.example.com/users/1")
	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// Check status code
	if resp.StatusCode == 404 {
		fmt.Println("User not found")
		return
	}

	if !resp.IsSuccess() {
		fmt.Printf("Request failed with status: %d\n", resp.StatusCode)
		return
	}

	// Get response headers
	contentType := resp.Header.Get("Content-Type")
	fmt.Printf("Content-Type: %s\n", contentType)

	// Read response body as string
	bodyStr := resp.ReadAllString()
	fmt.Printf("Response body: %s\n", bodyStr)

	// Read response body as bytes
	bodyBytes := resp.ReadAll()
	fmt.Printf("Response length: %d bytes\n", len(bodyBytes))
}
