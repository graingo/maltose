package mclient

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func ExampleBasicUsage() {
	// Create a default client
	client := New()

	// Send a GET request using the chain-style API
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		GET("https://api.example.com/users")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// Check response status using IsSuccess
	if !resp.IsSuccess() {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
		return
	}

	// Read response content
	bodyContent := resp.ReadAllString()
	fmt.Println("Response:", bodyContent)
}

func ExampleAdvancedConfiguration() {
	// Create a client with custom configuration
	client := NewWithConfig(ClientConfig{
		Timeout: 10 * time.Second,
		BaseURL: "https://api.example.com",
		Header: http.Header{
			"User-Agent": []string{"MaltoseClient/1.0"},
			"Accept":     []string{"application/json"},
		},
	})

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send a request with context
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

func ExampleJSONProcessing() {
	// Create a client
	client := New()

	// Define response structure
	type User struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		CreatedAt string `json:"created_at"`
	}

	// Method 1: Use SetResult for automatic JSON parsing
	var user User
	resp, err := client.R().
		SetResult(&user).
		GET("https://api.example.com/users/1")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// The result is automatically parsed into the user struct
	fmt.Printf("User: %s (%s)\n", user.Name, user.Email)

	// Method 2: Manual JSON parsing using resp.Parse
	resp, err = client.R().
		GET("https://api.example.com/users/2")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	// Check if request was successful using IsSuccess
	if !resp.IsSuccess() {
		log.Printf("Request returned non-success status: %d", resp.StatusCode)
		return
	}

	// Parse response directly using resp.Parse
	var anotherUser User
	if err := resp.Parse(&anotherUser); err != nil {
		log.Printf("JSON parsing failed: %v", err)
		return
	}

	fmt.Printf("Another user: %s (%s)\n", anotherUser.Name, anotherUser.Email)
}

func ExampleDataSending() {
	client := New()

	// 1. Sending JSON data
	userJSON := map[string]interface{}{
		"name":  "John Smith",
		"email": "john@example.com",
		"age":   30,
	}

	resp, err := client.R().
		SetBody(userJSON). // Automatically sets Content-Type to application/json
		POST("https://api.example.com/users")

	if err != nil {
		log.Printf("JSON request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("User created successfully, status code: %d\n", resp.StatusCode)

	// 2. Sending form data
	resp, err = client.R().
		SetForm("username", "john_smith").
		SetForm("password", "secret123").
		SetForm("remember", "true").
		POST("https://api.example.com/login")

	if err != nil {
		log.Printf("Form request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Login successful, status code: %d\n", resp.StatusCode)

	// 3. Sending form data using a map
	formData := map[string]string{
		"product_id": "12345",
		"quantity":   "2",
		"color":      "blue",
		"size":       "medium",
	}

	resp, err = client.R().
		SetFormMap(formData).
		POST("https://api.example.com/cart")

	if err != nil {
		log.Printf("Form map request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Added to cart, status code: %d\n", resp.StatusCode)
}

func ExampleErrorHandling() {
	client := New()

	// Define success and error response structures
	type SuccessResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}

	type ErrorResponse struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
		Code    int    `json:"code"`
	}

	// Prepare request parameters
	var successResp SuccessResponse
	var errorResp ErrorResponse

	resp, err := client.R().
		SetResult(&successResp). // Success response will be parsed here
		SetError(&errorResp).    // Error response will be parsed here
		GET("https://api.example.com/users/1")

	// First handle network errors
	if err != nil {
		log.Printf("Network error: %v", err)
		return
	}
	defer resp.Close()

	// Use IsSuccess to check if the request was successful
	if !resp.IsSuccess() {
		// The error response has been automatically parsed into errorResp
		fmt.Printf("API error: %s (code: %d)\n", errorResp.Error, errorResp.Code)
		return
	}

	// Success response has been automatically parsed into successResp
	fmt.Printf("Request successful: %s\n", successResp.Message)
	if data, ok := successResp.Data.(map[string]interface{}); ok {
		fmt.Printf("Username: %v\n", data["name"])
	}

	// Demonstration of how SetError works:
	// 1. When status code is 2xx, response is parsed into successResp
	// 2. When status code is not 2xx, response is parsed into errorResp
	// This means you don't need to manually check the status code and
	// parse the response differently - it's done automatically based
	// on the HTTP status code.
}

func ExampleMiddleware() {
	client := New()

	// Add a global request logging middleware
	client.Use(func(next HandlerFunc) HandlerFunc {
		return func(req *http.Request) (*http.Response, error) {
			startTime := time.Now()
			log.Printf("Starting request: %s %s", req.Method, req.URL.String())

			// Call the next handler
			resp, err := next(req)

			// Log duration and result
			duration := time.Since(startTime)
			if err != nil {
				log.Printf("Request failed: %v, duration: %v", err, duration)
			} else {
				log.Printf("Request completed: %s %s, status code: %d, duration: %v",
					req.Method, req.URL.String(), resp.StatusCode, duration)
			}

			return resp, err
		}
	})

	// Add a request ID middleware
	requestIDMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(req *http.Request) (*http.Response, error) {
			// Set request ID
			req.Header.Set("X-Request-ID", fmt.Sprintf("req-%d", time.Now().UnixNano()))
			return next(req)
		}
	}

	// Create request, using specific middleware
	resp, err := client.R().
		Use(requestIDMiddleware). // Middleware only for this request
		GET("https://api.example.com/status")

	if err != nil {
		log.Printf("Request failed: %v", err)
		return
	}
	defer resp.Close()

	fmt.Printf("Request status: %d\n", resp.StatusCode)
}

func ExampleRealWorldScenario() {
	// Create a production-ready client
	client := NewWithConfig(ClientConfig{
		Timeout: 30 * time.Second,
		BaseURL: "https://api.github.com",
		Header: http.Header{
			"Accept":     []string{"application/json"},
			"User-Agent": []string{"MaltoseClient-Example/1.0"},
		},
	})

	// Define API response types
	type GithubUser struct {
		Login     string `json:"login"`
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Company   string `json:"company"`
		Blog      string `json:"blog"`
		Location  string `json:"location"`
		Email     string `json:"email"`
		Followers int    `json:"followers"`
		Following int    `json:"following"`
	}

	type GithubRepo struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
		Stars       int    `json:"stargazers_count"`
		Forks       int    `json:"forks_count"`
		Language    string `json:"language"`
	}

	type GithubErrorResponse struct {
		Message          string `json:"message"`
		DocumentationURL string `json:"documentation_url"`
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user information
	var user GithubUser
	var errorResp GithubErrorResponse

	resp, err := client.R().
		SetContext(ctx).
		SetResult(&user).
		SetError(&errorResp).
		GET("/users/octocat")

	if err != nil {
		log.Printf("Network error: %v", err)
		return
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Printf("API error: %s (docs: %s)",
			errorResp.Message,
			errorResp.DocumentationURL)
		return
	}

	fmt.Printf("User: %s (%s)\n", user.Name, user.Login)
	fmt.Printf("Company: %s, Location: %s\n", user.Company, user.Location)
	fmt.Printf("Followers: %d, Following: %d\n", user.Followers, user.Following)

	// Get user's repositories
	var repos []GithubRepo
	resp, err = client.R().
		SetContext(ctx).
		SetQueryMap(map[string]string{
			"sort":      "updated",
			"direction": "desc",
			"per_page":  "5",
		}).
		SetResult(&repos).
		SetError(&errorResp).
		GET(fmt.Sprintf("/users/%s/repos", user.Login))

	if err != nil {
		log.Printf("Repository fetch failed: %v", err)
		return
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Printf("Repository API error: %s", errorResp.Message)
		return
	}

	fmt.Printf("\nRecent repositories:\n")
	for i, repo := range repos {
		fmt.Printf("%d. %s - %s (Language: %s, Stars: %d)\n",
			i+1, repo.Name, repo.Description, repo.Language, repo.Stars)
	}
}

func ExampleChainedCalls() {
	// Create a client with configuration
	client := NewWithConfig(ClientConfig{
		Timeout: 10 * time.Second,
		BaseURL: "https://api.example.com",
		Header: http.Header{
			"User-Agent": []string{"MaltoseClient/1.0"},
		},
	})

	// Define various response types
	type Product struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
		InStock     bool    `json:"in_stock"`
	}

	type ProductResponse struct {
		Data Product `json:"data"`
	}

	type ProductListResponse struct {
		Total    int       `json:"total"`
		Page     int       `json:"page"`
		PageSize int       `json:"page_size"`
		Products []Product `json:"products"`
	}

	type ErrorResponse struct {
		Error   string `json:"error"`
		Code    int    `json:"code"`
		Details string `json:"details"`
	}

	// 1. Get product list
	var productList ProductListResponse
	var errorResp ErrorResponse

	resp, err := client.R().
		SetContext(context.Background()).
		SetHeaders(map[string]string{
			"Accept":        "application/json",
			"Cache-Control": "no-cache",
		}).
		SetQueryMap(map[string]string{
			"category": "electronics",
			"sort":     "price",
			"order":    "asc",
			"page":     "1",
			"limit":    "10",
		}).
		SetResult(&productList).
		SetError(&errorResp).
		GET("/products")

	if err != nil {
		log.Printf("Product list fetch failed: %v", err)
		return
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Printf("API error: %s (code: %d)", errorResp.Error, errorResp.Code)
		return
	}

	fmt.Printf("Found %d products, total: %d\n", len(productList.Products), productList.Total)
	for i, product := range productList.Products {
		fmt.Printf("%d. %s - $%.2f (%s)\n", i+1, product.Name, product.Price, product.Category)
	}

	// 2. Get a single product detail
	var productResp ProductResponse
	resp, err = client.R().
		SetContext(context.Background()).
		SetResult(&productResp).
		SetError(&errorResp).
		GET("/products/42")

	if err != nil {
		log.Printf("Product detail fetch failed: %v", err)
		return
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Printf("API error: %s (code: %d)", errorResp.Error, errorResp.Code)
		return
	}

	product := productResp.Data
	fmt.Printf("\nProduct details:\n")
	fmt.Printf("Name: %s\n", product.Name)
	fmt.Printf("Description: %s\n", product.Description)
	fmt.Printf("Price: $%.2f\n", product.Price)
	fmt.Printf("In stock: %v\n", product.InStock)

	// 3. Update product info
	updateData := map[string]interface{}{
		"price":    product.Price * 1.1, // 10% price increase
		"in_stock": true,
	}

	var updatedProduct Product
	resp, err = client.R().
		SetContext(context.Background()).
		SetHeader("Content-Type", "application/json").
		SetBody(updateData).
		SetResult(&updatedProduct).
		SetError(&errorResp).
		PUT(fmt.Sprintf("/products/%d", product.ID))

	if err != nil {
		log.Printf("Product update failed: %v", err)
		return
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Printf("API error: %s (code: %d)", errorResp.Error, errorResp.Code)
		return
	}

	fmt.Printf("\nProduct updated:\n")
	fmt.Printf("New price: $%.2f\n", updatedProduct.Price)
	fmt.Printf("Stock status: %v\n", updatedProduct.InStock)
}
