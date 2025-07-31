package mhttp_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Controller for Middleware ---

type TestMiddlewareController struct{}

type SuccessReq struct {
	mmeta.Meta `path:"/success" method:"get"`
}
type SuccessRes struct {
	Data string `json:"data"`
}

func (c *TestMiddlewareController) GetSuccess(_ context.Context, _ *SuccessReq) (*SuccessRes, error) {
	return &SuccessRes{Data: "ok"}, nil
}

type ErrorReq struct {
	mmeta.Meta `path:"/error" method:"get"`
}
type ErrorRes struct{}

func (c *TestMiddlewareController) GetError(_ context.Context, _ *ErrorReq) (*ErrorRes, error) {
	return nil, merror.New("something went wrong")
}

// --- Tests ---

func TestMiddleware(t *testing.T) {
	t.Run("execution_order", func(t *testing.T) {
		var results []string
		var mu sync.Mutex

		teardown := setupServer(t, func(s *mhttp.Server) {
			s.Use(func(r *mhttp.Request) {
				mu.Lock()
				results = append(results, "m1-in")
				mu.Unlock()
				r.Next()
				mu.Lock()
				results = append(results, "m1-out")
				mu.Unlock()
			}, func(r *mhttp.Request) {
				mu.Lock()
				results = append(results, "m2-in")
				mu.Unlock()
				r.Next()
				mu.Lock()
				results = append(results, "m2-out")
				mu.Unlock()
			})
			s.GET("/test", func(r *mhttp.Request) {
				mu.Lock()
				results = append(results, "handler")
				mu.Unlock()
				r.String(http.StatusOK, "ok")
			})
		})
		defer teardown()

		_, err := http.Get(baseURL + "/test")
		require.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()
		expected := []string{"m1-in", "m2-in", "handler", "m2-out", "m1-out"}
		assert.Equal(t, expected, results)
	})

	t.Run("grouping", func(t *testing.T) {
		var groupMiddlewareCalled bool
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.GET("/no-group", func(r *mhttp.Request) {
				r.String(http.StatusOK, "ok")
			})

			v1 := s.Group("/v1")
			v1.Use([]mhttp.MiddlewareFunc{func(r *mhttp.Request) {
				groupMiddlewareCalled = true
				r.Next()
			}})
			v1.GET("/in-group", func(r *mhttp.Request) {
				r.String(http.StatusOK, "ok")
			})
		})
		defer teardown()

		_, err := http.Get(baseURL + "/v1/in-group")
		require.NoError(t, err)
		assert.True(t, groupMiddlewareCalled, "Group middleware should be called")

		groupMiddlewareCalled = false // reset
		_, err = http.Get(baseURL + "/no-group")
		require.NoError(t, err)
		assert.False(t, groupMiddlewareCalled, "Group middleware should not be called")
	})

	t.Run("single_route", func(t *testing.T) {
		var routeMiddlewareCalled bool
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.GET("/test", func(r *mhttp.Request) {
				r.String(http.StatusOK, "ok")
			}, func(r *mhttp.Request) {
				routeMiddlewareCalled = true
				r.Next()
			})
		})
		defer teardown()

		_, err := http.Get(baseURL + "/test")
		require.NoError(t, err)
		assert.True(t, routeMiddlewareCalled)
	})

	t.Run("standard_response", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.Use(mhttp.MiddlewareResponse())
			s.Bind(&TestMiddlewareController{})
		})
		defer teardown()

		t.Run("success_case", func(t *testing.T) {
			resp, err := http.Get(baseURL + "/success")
			require.NoError(t, err)
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			var respMap map[string]interface{}
			err = json.Unmarshal(body, &respMap)
			require.NoError(t, err)

			assert.Equal(t, float64(0), respMap["code"])
			assert.Equal(t, "OK", respMap["message"])
		})

		t.Run("error_case", func(t *testing.T) {
			resp, err := http.Get(baseURL + "/error")
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			var respMap map[string]interface{}
			err = json.Unmarshal(body, &respMap)
			require.NoError(t, err)

			assert.Equal(t, float64(2000), respMap["code"]) // internal error code
			assert.Equal(t, "something went wrong", respMap["message"])
			assert.Nil(t, respMap["data"])
		})
	})

	t.Run("abort", func(t *testing.T) {
		var handlerCalled bool
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.Use(func(r *mhttp.Request) {
				r.String(http.StatusUnauthorized, "aborted")
				r.Abort()
			})
			s.GET("/test", func(_ *mhttp.Request) {
				handlerCalled = true
			})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/test")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, "aborted", string(body))
		assert.False(t, handlerCalled, "Handler should not be called after abort")
	})

	t.Run("rate_limit", func(t *testing.T) {
		config := mhttp.RateLimitConfig{
			Rate:  5, // 5 requests per second
			Burst: 1, // Burst of 1
		}

		teardown := setupServer(t, func(s *mhttp.Server) {
			s.Use(mhttp.MiddlewareRateLimit(config))
			s.GET("/limited", func(r *mhttp.Request) {
				r.String(http.StatusOK, "ok")
			})
		})
		defer teardown()

		// First request should pass
		resp, err := http.Get(baseURL + "/limited")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Immediate second request should be rate-limited
		resp, err = http.Get(baseURL + "/limited")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

		// Wait for token refill
		time.Sleep(500 * time.Millisecond)

		// Third request should pass again
		resp, err = http.Get(baseURL + "/limited")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("recovery", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			// Recovery middleware is added by default in New()
			s.GET("/panic", func(_ *mhttp.Request) {
				panic("something went terribly wrong")
			})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/panic")
		require.NoError(t, err)
		defer resp.Body.Close()

		// The default recovery middleware writes a plain text response.
		// The MiddlewareResponse would format it as JSON.
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Internal Panic")
	})

	t.Run("recovery_with_custom_handler", func(t *testing.T) {
		var customHandlerCalled bool
		var receivedError error

		teardown := setupServer(t, func(s *mhttp.Server) {
			// Set custom panic handler
			s.WithPanicHandler(func(r *mhttp.Request, err error) {
				customHandlerCalled = true
				receivedError = err
				r.JSON(http.StatusTeapot, map[string]string{
					"error":   "custom panic handler",
					"message": err.Error(),
				})
			})

			s.GET("/panic", func(_ *mhttp.Request) {
				panic("custom panic test")
			})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/panic")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Verify custom handler was called
		assert.True(t, customHandlerCalled, "Custom panic handler should be called")
		assert.NotNil(t, receivedError, "Error should be passed to custom handler")
		assert.Contains(t, receivedError.Error(), "custom panic test")

		// Verify custom response
		assert.Equal(t, http.StatusTeapot, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		var respMap map[string]interface{}
		err = json.Unmarshal(body, &respMap)
		require.NoError(t, err)

		assert.Equal(t, "custom panic handler", respMap["error"])
		assert.Contains(t, respMap["message"], "custom panic test")
	})
}
