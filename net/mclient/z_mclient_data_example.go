package mclient

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ExampleProcessJSON demonstrates how to process JSON data
func ExampleProcessJSON() {
	// Create client
	client := New()

	// Define response structure
	type User struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}

	// Prepare result container
	var user User

	// Send request and parse response in one step
	resp, err := client.R().
		SetResult(&user). // Set result container for automatic parsing
		GET("https://api.example.com/users/1")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Printf("Error response: %d", resp.StatusCode)
		return
	}

	// Access the parsed data
	fmt.Printf("User: %s (%s)\n", user.Name, user.Email)

	// Alternative manual parsing
	var anotherUser User
	err = json.Unmarshal(resp.ReadAll(), &anotherUser)
	if err != nil {
		log.Printf("Failed to parse JSON: %v", err)
		return
	}

	// Access manually parsed data
	fmt.Printf("User ID: %d\n", anotherUser.ID)
}

// ExampleSendData demonstrates how to send different types of data
func ExampleSendData() {
	// Create client
	client := New()

	// Send JSON data
	jsonResp, err := client.R().
		SetBody(map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}).
		POST("https://api.example.com/users")

	if err != nil {
		log.Printf("JSON request failed: %v", err)
	} else {
		defer jsonResp.Close()
		fmt.Printf("JSON request status: %d\n", jsonResp.StatusCode)
	}

	// Send form data
	formResp, err := client.R().
		SetFormMap(map[string]string{
			"username": "johndoe",
			"password": "securepassword",
		}).
		POST("https://api.example.com/login")

	if err != nil {
		log.Printf("Form request failed: %v", err)
	} else {
		defer formResp.Close()
		fmt.Printf("Form request status: %d\n", formResp.StatusCode)
	}

	// Send query parameters
	queryResp, err := client.R().
		SetQuery("category", "books").
		SetQuery("sort", "price").
		SetQuery("order", "asc").
		GET("https://api.example.com/products")

	if err != nil {
		log.Printf("Query request failed: %v", err)
	} else {
		defer queryResp.Close()
		fmt.Printf("Query request status: %d\n", queryResp.StatusCode)
	}
}

// ExampleHandleError demonstrates how to handle error responses
func ExampleHandleError() {
	// Create client
	client := New()

	// Define error response structure
	type ErrorResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Details string `json:"details,omitempty"`
	}

	// Prepare error container
	var errorResponse ErrorResponse

	// Send request
	resp, err := client.R().
		SetError(&errorResponse). // Set error container for automatic parsing
		GET("https://api.example.com/users/999")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// Check if request was successful
	if resp.IsSuccess() {
		fmt.Println("Request successful")
		return
	}

	// Handle different error status codes
	switch resp.StatusCode {
	case 404:
		fmt.Printf("Resource not found: %s\n", errorResponse.Message)
	case 401, 403:
		fmt.Printf("Authentication error: %s\n", errorResponse.Message)
	case 400:
		fmt.Printf("Bad request: %s (%s)\n", errorResponse.Message, errorResponse.Details)
	default:
		fmt.Printf("Unexpected error: %s (Code: %d)\n", errorResponse.Message, errorResponse.Code)
	}
}

// ExampleCustomDataTypes demonstrates how to handle custom data types
func ExampleCustomDataTypes() {
	// Create client
	client := New()

	// Define custom time format
	type CustomTime time.Time

	// Define custom unmarshaler
	type Product struct {
		ID        int        `json:"id"`
		Name      string     `json:"name"`
		Price     float64    `json:"price"`
		CreatedAt CustomTime `json:"created_at"`
	}

	// Prepare result container
	var product Product

	// Send request
	resp, err := client.R().
		SetResult(&product).
		GET("https://api.example.com/products/1")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Printf("Error response: %d", resp.StatusCode)
		return
	}

	// Access the parsed data
	fmt.Printf("Product: %s (%.2f)\n", product.Name, product.Price)
	fmt.Printf("Created: %v\n", time.Time(product.CreatedAt))
}
