package mclient_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/graingo/maltose/net/mclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Basic Request Tests
// -----------------------------------------------------------------------------

// TestBasicRequest tests basic request functionality
func TestBasicRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert request method
		assert.Equal(t, "GET", r.Method, "Expected GET request")

		// Assert headers
		assert.Equal(t, "application/json", r.Header.Get("Accept"), "Expected Accept: application/json header")

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"success"}`))
	}))
	defer server.Close()

	// Create client
	client := mclient.New()

	// Send request
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		GET(server.URL)

	// Assert response
	require.NoError(t, err, "Should not return error")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse response
	var result map[string]string
	err = resp.Parse(&result)
	require.NoError(t, err, "Should parse response without error")

	// Assert response body
	assert.Equal(t, "success", result["message"], "Expected message 'success'")
}

// TestJSONBodyRequest tests sending JSON request bodies
func TestJSONBodyRequest(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert content type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Expected Content-Type: application/json")

		// Read request body
		var requestBody map[string]interface{}
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		// Validate request body
		err := json.Unmarshal(body, &requestBody)
		require.NoError(t, err, "Should parse request body without error")

		// Assert request body values
		assert.Equal(t, "John Doe", requestBody["name"], "Expected name 'John Doe'")

		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 123, "name": "John Doe", "status": "created"}`))
	}))
	defer server.Close()

	// Create client
	client := mclient.New()

	// Prepare request data
	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	// Define response structure
	type CreateResponse struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	// Prepare result container
	var result CreateResponse

	// Send request
	resp, err := client.R().
		SetBody(data).
		SetResult(&result).
		POST(server.URL)

	// Assert response
	require.NoError(t, err, "Should not return error")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status code 201")

	// Assert parsed result
	assert.Equal(t, 123, result.ID, "Expected ID 123")
	assert.Equal(t, "John Doe", result.Name, "Expected name 'John Doe'")
	assert.Equal(t, "created", result.Status, "Expected status 'created'")
}
