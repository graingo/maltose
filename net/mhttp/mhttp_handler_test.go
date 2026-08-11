package mhttp_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerSupportsHTTPTest(t *testing.T) {
	server := mhttp.New()
	server.GET("/ping", func(request *mhttp.Request) {
		request.String(http.StatusOK, "pong")
	})

	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "pong", response.Body.String())
}

func TestStartListenerUsesProvidedListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := mhttp.New()
	require.NoError(t, server.SetConfigWithMap(map[string]any{"graceful_wait_time": 0}))
	server.GET("/listener", func(request *mhttp.Request) {
		request.String(http.StatusOK, "ready")
	})

	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.StartListener(context.Background(), listener)
	}()

	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://" + listener.Addr().String() + "/listener")
	require.NoError(t, err)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	assert.Equal(t, "ready", string(body))

	require.NoError(t, server.Stop(context.Background()))
	select {
	case err := <-serveDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("StartListener did not return after Stop")
	}
}

func TestStopClosesImmediatelyWhenGracefulShutdownIsDisabled(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := mhttp.New()
	require.NoError(t, server.SetConfigWithMap(map[string]any{"graceful_enable": false}))
	server.GET("/ready", func(request *mhttp.Request) {
		request.String(http.StatusOK, "ready")
	})
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.StartListener(context.Background(), listener)
	}()

	response, err := http.Get("http://" + listener.Addr().String() + "/ready")
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.NoError(t, server.Stop(context.Background()))
	select {
	case err := <-serveDone:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("server did not stop after graceful shutdown was disabled")
	}
}
