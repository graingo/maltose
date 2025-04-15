package mclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/graingo/maltose/net/mclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetryMechanism tests the request retry functionality
func TestRetryMechanism(t *testing.T) {
	// Track number of requests
	requestCount := 0

	// Create test server that fails the first 2 requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Fail the first 2 requests
		if requestCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Succeed on the 3rd request
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`success after retry`))
	}))
	defer server.Close()

	// Create client
	client := mclient.New()

	// Send request with retry configuration
	resp, err := client.R().
		SetRetrySimple(3, 10*time.Millisecond). // 3 retries, 10ms interval
		GET(server.URL)

	// Assert response
	require.NoError(t, err, "Should not return error")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Assert total number of requests (initial + 2 retries = 3)
	assert.Equal(t, 3, requestCount, "Expected 3 total requests")
}

// TestContextTimeout tests request timing out via context
func TestContextTimeout(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the context timeout
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client
	client := mclient.New()

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Send request with short timeout context
	_, err := client.R().
		SetContext(ctx).
		GET(server.URL)

	// Assert that context deadline caused an error
	assert.Error(t, err, "Should return timeout error")
}
