package mclient_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mclient"
	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	t.Run("auth_middleware", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := mclient.New()
		client.Use(mclient.MiddlewareFunc(func(next mclient.HandlerFunc) mclient.HandlerFunc {
			return func(req *mclient.Request) (*mclient.Response, error) {
				req.SetHeader("Authorization", "Bearer test-token")
				return next(req)
			}
		}))

		resp, err := client.R().Get(server.URL)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("rate_limit_middleware", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := mclient.New()
		client.Use(mclient.MiddlewareRateLimit(mclient.RateLimitConfig{
			RequestsPerSecond: 2,
			Burst:             1,
		}))

		startTime := time.Now()
		for i := 0; i < 3; i++ {
			resp, err := client.R().Get(server.URL)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
		duration := time.Since(startTime)

		// The first request uses the burst token; the next two wait about 500 ms each.
		assert.Greater(t, duration, 900*time.Millisecond, "Expected total time to be > 900ms")
	})

	t.Run("recovery_middleware", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := mclient.New()
		client.Use(func(_ mclient.HandlerFunc) mclient.HandlerFunc {
			return func(_ *mclient.Request) (*mclient.Response, error) {
				return nil, merror.New("middleware error")
			}
		})

		_, err := client.R().Get(server.URL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "middleware error")
	})

	t.Run("request_level_middleware", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "client", r.Header.Get("X-Middleware-Scope"))
			assert.Equal(t, "request", r.Header.Get("X-Request-ID"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := mclient.New()
		client.Use(func(next mclient.HandlerFunc) mclient.HandlerFunc {
			return func(req *mclient.Request) (*mclient.Response, error) {
				req.SetHeader("X-Middleware-Scope", "client")
				return next(req)
			}
		})

		requestMiddleware := func(next mclient.HandlerFunc) mclient.HandlerFunc {
			return func(req *mclient.Request) (*mclient.Response, error) {
				req.SetHeader("X-Request-ID", "request")
				return next(req)
			}
		}

		_, err := client.R().Use(requestMiddleware).Get(server.URL)
		require.NoError(t, err)
	})

	t.Run("log_middleware", func(t *testing.T) {
		previousBodyLimit := mclient.LogMaxBodySize
		mclient.LogMaxBodySize = -1
		defer func() { mclient.LogMaxBodySize = previousBodyLimit }()

		var buf bytes.Buffer
		cfg := mlog.Config{
			Writer: &buf,
			Level:  mlog.DebugLevel,
			Format: "json",
		}
		logger := mlog.New(&cfg)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/error" {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("server error"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		}))
		defer server.Close()

		client := mclient.New()
		client.Use(mclient.MiddlewareLog(logger))

		_, err := client.R().SetBody("req_body").Post(server.URL + "?token=super-secret&view=full")
		require.NoError(t, err)

		logStr := buf.String()
		assert.Contains(t, logStr, "http client request started")
		assert.Contains(t, logStr, "http client request finished")
		assert.Contains(t, logStr, `"method":"POST"`)
		assert.Contains(t, logStr, `"request_body":"req_body"`)
		assert.Contains(t, logStr, `"status":200`)
		assert.Contains(t, logStr, `"response_body":"success"`)
		assert.NotContains(t, logStr, "super-secret")
		assert.Contains(t, logStr, `"query_keys":["token","view"]`)

		buf.Reset()
		_, err = client.R().Get(server.URL + "/error")
		require.NoError(t, err)

		logStr = buf.String()
		assert.Contains(t, logStr, "http client request finished with error status")
		assert.Contains(t, logStr, `"status":500`)
		assert.Contains(t, logStr, `"response_body":"server error"`)

		clientWithNilLogger := mclient.New()
		clientWithNilLogger.Use(mclient.MiddlewareLog(nil))
		assert.NotPanics(t, func() {
			_, _ = clientWithNilLogger.R().Get(server.URL)
		})
	})

	t.Run("nil_response_from_middleware", func(t *testing.T) {
		client := mclient.New()
		client.Use(func(_ mclient.HandlerFunc) mclient.HandlerFunc {
			return func(*mclient.Request) (*mclient.Response, error) { return nil, nil }
		})
		_, err := client.R().Get("http://example.invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil response")
	})
}
