package mclient_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/graingo/maltose/net/mclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Client Configuration Tests
// -----------------------------------------------------------------------------

// TestCloneClient tests cloning a client with its configuration
func TestCloneClient(t *testing.T) {
	// Create original client with custom configuration
	originalClient := mclient.New()
	originalClient.SetConfig(mclient.ClientConfig{
		BaseURL: "https://example.com",
		Header: http.Header{
			"User-Agent": []string{"TestAgent"},
		},
	})

	// Add middleware to original client
	middlewareExecuted := false
	originalClient.Use(func(next mclient.HandlerFunc) mclient.HandlerFunc {
		return func(req *mclient.Request) (*mclient.Response, error) {
			middlewareExecuted = true
			return next(req)
		}
	})

	// Clone the client
	clonedClient := originalClient.Clone()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that headers from original client are preserved
		assert.Equal(t, "TestAgent", r.Header.Get("User-Agent"), "Expected User-Agent: TestAgent")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Send request with cloned client
	resp, err := clonedClient.R().GET(server.URL)

	// Assert response
	require.NoError(t, err, "Should not return error")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Assert middleware was executed
	assert.True(t, middlewareExecuted, "Expected middleware to be executed after client cloning")
}
