package mhttp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var baseURL string

// setupServer creates an isolated test server and returns its cleanup function.
func setupServer(t *testing.T, configurator func(s *mhttp.Server)) func() {
	t.Helper()
	s := mhttp.New()
	require.NoError(t, s.SetConfigWithMap(map[string]any{"healthCheck": ""}))

	if configurator != nil {
		configurator(s)
	}

	testServer := httptest.NewServer(s.Handler())
	baseURL = testServer.URL
	return testServer.Close
}

func TestServer(t *testing.T) {
	t.Run("new_server", func(t *testing.T) {
		s := mhttp.New()
		assert.NotNil(t, s, "New() should not return nil")
	})

	t.Run("basic_route", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.GET("/ping", func(r *mhttp.Request) {
				r.String(http.StatusOK, "pong")
			})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/ping")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "pong", string(body))
	})
}
