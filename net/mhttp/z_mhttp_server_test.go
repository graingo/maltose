package mhttp_test

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testHost = "127.0.0.1:18080"
	baseURL  = "http://" + testHost
)

// setupServer creates a new server instance, runs it in a goroutine,
// and waits for it to be ready. It returns the server instance and a teardown function.
func setupServer(t *testing.T, configurator func(s *mhttp.Server)) func() {
	s := mhttp.New()
	s.SetAddress(testHost)

	// Disable health check for some tests to avoid conflicts.
	// It can be re-enabled by the configurator.
	s.SetConfigWithMap(map[string]any{
		"healthCheck": false,
	})

	if configurator != nil {
		configurator(s)
	}

	serverCtx, serverCancel := context.WithCancel(context.Background())

	go func() {
		if err := s.Start(serverCtx); err != nil && err != http.ErrServerClosed {
			t.Logf("Server failed to start: %v", err)
		}
	}()

	// Wait for the server to be ready.
	waitForServerReady(t, s)

	return func() {
		serverCancel()
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer stopCancel()
		err := s.Stop(stopCtx)
		assert.NoError(t, err, "Server should stop gracefully")
		// Add a small delay to ensure the port is released
		time.Sleep(100 * time.Millisecond)
	}
}

// waitForServerReady polls the server until it's responsive.
func waitForServerReady(t *testing.T, s *mhttp.Server) {
	// Use a TCP dial check which is more reliable and has no side effects.
	for i := 0; i < 25; i++ { // try for 5 seconds
		conn, err := net.DialTimeout("tcp", testHost, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("Server at %s did not become ready", baseURL)
}

func TestServer_New(t *testing.T) {
	s := mhttp.New()
	assert.NotNil(t, s, "New() should not return nil")
}

func TestServer_BasicRoute(t *testing.T) {
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
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "pong", string(body))
}

func TestServer_HealthCheck(t *testing.T) {
	teardown := setupServer(t, func(s *mhttp.Server) {
		// Explicitly enable health check for this test
		s.SetConfigWithMap(map[string]any{
			"healthCheck": true,
		})
	})
	defer teardown()

	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"status":"ok"`)
}

func TestServer_StaticFile(t *testing.T) {
	// Create a dummy file to serve
	tmpDir := t.TempDir()
	filePath := tmpDir + "/test.txt"
	fileContent := "hello world"
	err := ioutil.WriteFile(filePath, []byte(fileContent), 0644)
	require.NoError(t, err)

	teardown := setupServer(t, func(s *mhttp.Server) {
		s.SetStaticPath("/static", tmpDir)
	})
	defer teardown()

	resp, err := http.Get(baseURL + "/static/test.txt")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, fileContent, string(body))
}
