package mhttp_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareLogCapturesBoundedBodiesWithoutQueryValues(t *testing.T) {
	previousBodyLimit := mhttp.LogMaxBodySize
	mhttp.LogMaxBodySize = 4
	t.Cleanup(func() { mhttp.LogMaxBodySize = previousBodyLimit })

	var output bytes.Buffer
	logger := mlog.New(&mlog.Config{
		Writer: &output,
		Level:  mlog.DebugLevel,
		Format: "json",
	})
	t.Cleanup(func() { require.NoError(t, logger.Close()) })

	server := mhttp.New(&mhttp.Config{
		HealthCheck: "/health",
		Logger:      logger,
	})
	server.Use(mhttp.MiddlewareLog())
	server.POST("/logged", func(request *mhttp.Request) {
		body, err := io.ReadAll(request.Request.Body)
		require.NoError(t, err)
		assert.Equal(t, "request-body", string(body))
		request.String(http.StatusCreated, "response-body")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/logged?token=secret&view=full",
		strings.NewReader("request-body"),
	)
	server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
	logOutput := output.String()
	assert.Contains(t, logOutput, "http server request started")
	assert.Contains(t, logOutput, "http server request finished")
	assert.Contains(t, logOutput, `"request_body":"requ..."`)
	assert.Contains(t, logOutput, `"response_body":"resp..."`)
	assert.Contains(t, logOutput, `"query_keys":["token","view"]`)
	assert.NotContains(t, logOutput, "secret")
	assert.NotContains(t, logOutput, "full")

	output.Reset()
	healthRecorder := httptest.NewRecorder()
	server.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	assert.Equal(t, http.StatusOK, healthRecorder.Code)
	assert.Empty(t, output.String())
}

func TestMiddlewareRateLimitSupportsSkipAndCustomResponses(t *testing.T) {
	config := mhttp.RateLimitConfig{
		Rate:  0.001,
		Burst: 1,
		SkipFunc: func(request *mhttp.Request) bool {
			return request.Request.URL.Path == "/skip"
		},
		ErrorHandler: func(request *mhttp.Request) {
			request.String(http.StatusTeapot, "rate limited")
		},
	}

	server := mhttp.New()
	server.Use(mhttp.MiddlewareRateLimit(config))
	server.GET("/limited", func(request *mhttp.Request) {
		request.String(http.StatusOK, "ok")
	})
	server.GET("/skip", func(request *mhttp.Request) {
		request.String(http.StatusOK, "skipped")
	})

	first := httptest.NewRecorder()
	server.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/limited", nil))
	assert.Equal(t, http.StatusOK, first.Code)

	second := httptest.NewRecorder()
	server.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/limited", nil))
	assert.Equal(t, http.StatusTeapot, second.Code)
	assert.Equal(t, "rate limited", second.Body.String())

	skipped := httptest.NewRecorder()
	server.ServeHTTP(skipped, httptest.NewRequest(http.MethodGet, "/skip", nil))
	assert.Equal(t, http.StatusOK, skipped.Code)
	assert.Equal(t, "skipped", skipped.Body.String())
}

func TestMiddlewareRateLimitByIPIsolatesClients(t *testing.T) {
	server := mhttp.New()
	server.Use(mhttp.MiddlewareRateLimitByIP(mhttp.RateLimitConfig{
		Rate:  0.001,
		Burst: 1,
	}))
	server.GET("/limited", func(request *mhttp.Request) {
		request.String(http.StatusOK, "ok")
	})

	request := func(ip string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(http.MethodGet, "/limited", nil)
		httpRequest.Header.Set("X-Forwarded-For", ip)
		server.ServeHTTP(recorder, httpRequest)
		return recorder
	}

	assert.Equal(t, http.StatusOK, request("192.0.2.1").Code)
	assert.Equal(t, http.StatusTooManyRequests, request("192.0.2.1").Code)
	assert.Equal(t, http.StatusOK, request("192.0.2.2").Code)
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := mhttp.DefaultRateLimitConfig()
	assert.Equal(t, float64(100), config.Rate)
	assert.Equal(t, 10, config.Burst)
}
