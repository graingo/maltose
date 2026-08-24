package mclient_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/graingo/maltose/net/mclient"
)

// Example demonstrates a basic request with the chain-style API.
func Example() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	response, err := mclient.New().R().
		SetHeader("Accept", "application/json").
		Get(server.URL)
	if err != nil {
		fmt.Println("request failed")
		return
	}
	defer response.Close()

	fmt.Println(response.StatusCode)
	// Output:
	// 200
}

// Example_jSON demonstrates JSON request and response handling.
func Example_jSON() {
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	type CreateResponse struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":7,"name":"John Doe","status":"active"}`))
	}))
	defer server.Close()

	var result CreateResponse
	response, err := mclient.New().R().
		SetBody(User{Name: "John Doe", Email: "john@example.com"}).
		SetResult(&result).
		Post(server.URL)
	if err != nil {
		fmt.Println("request failed")
		return
	}
	defer response.Close()

	fmt.Println(result.ID, result.Name, result.Status)
	// Output:
	// 7 John Doe active
}

// Example_retry demonstrates retrying temporary server failures.
func Example_retry() {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	config := mclient.RetryConfig{
		Count:         3,
		BaseInterval:  time.Millisecond,
		MaxInterval:   time.Millisecond,
		BackoffFactor: 1,
	}
	response, err := mclient.New().R().SetRetry(config).Get(server.URL)
	if err != nil {
		fmt.Println("request failed")
		return
	}
	defer response.Close()

	fmt.Println(response.StatusCode, attempts.Load())
	// Output:
	// 204 3
}

// Example_customRetryCondition demonstrates selecting retryable responses.
func Example_customRetryCondition() {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	retryRateLimits := func(response *http.Response, err error) bool {
		return err == nil && response != nil && response.StatusCode == http.StatusTooManyRequests
	}
	config := mclient.RetryConfig{
		Count:         3,
		BaseInterval:  time.Millisecond,
		MaxInterval:   time.Millisecond,
		BackoffFactor: 1,
	}
	response, err := mclient.New().R().
		SetRetry(config).
		SetRetryCondition(retryRateLimits).
		Get(server.URL)
	if err != nil {
		fmt.Println("request failed")
		return
	}
	defer response.Close()

	fmt.Println(response.StatusCode, attempts.Load())
	// Output:
	// 204 2
}

// Example_middleware demonstrates adding authentication with client middleware.
func Example_middleware() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := mclient.New()
	client.Use(func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(request *mclient.Request) (*mclient.Response, error) {
			request.SetHeader("Authorization", "Bearer test-token")
			return next(request)
		}
	})

	response, err := client.R().Get(server.URL)
	if err != nil {
		fmt.Println("request failed")
		return
	}
	defer response.Close()

	fmt.Println(response.StatusCode)
	// Output:
	// 204
}
