package mhttp_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

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

func (c *TestMiddlewareController) GetSuccess(ctx context.Context, req *SuccessReq) (*SuccessRes, error) {
	return &SuccessRes{Data: "ok"}, nil
}

type ErrorReq struct {
	mmeta.Meta `path:"/error" method:"get"`
}
type ErrorRes struct{}

func (c *TestMiddlewareController) GetError(ctx context.Context, req *ErrorReq) (*ErrorRes, error) {
	return nil, merror.New("something went wrong")
}

// --- Tests ---

func TestMiddleware_ExecutionOrder(t *testing.T) {
	var results []string
	teardown := setupServer(t, func(s *mhttp.Server) {
		s.Use(func(r *mhttp.Request) {
			results = append(results, "m1-in")
			r.Next()
			results = append(results, "m1-out")
		}, func(r *mhttp.Request) {
			results = append(results, "m2-in")
			r.Next()
			results = append(results, "m2-out")
		})
		s.GET("/test", func(r *mhttp.Request) {
			results = append(results, "handler")
			r.String(http.StatusOK, "ok")
		})
	})
	defer teardown()

	_, err := http.Get(baseURL + "/test")
	require.NoError(t, err)

	expected := []string{"m1-in", "m2-in", "handler", "m2-out", "m1-out"}
	assert.Equal(t, expected, results)
}

func TestMiddleware_Group(t *testing.T) {
	var groupMiddlewareCalled bool
	teardown := setupServer(t, func(s *mhttp.Server) {
		// Route without group middleware
		s.GET("/no-group", func(r *mhttp.Request) {
			r.String(http.StatusOK, "ok")
		})

		// Group with middleware
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

	// Test route in group
	_, err := http.Get(baseURL + "/v1/in-group")
	require.NoError(t, err)
	assert.True(t, groupMiddlewareCalled, "Group middleware should be called")

	// Test route not in group
	groupMiddlewareCalled = false // reset
	_, err = http.Get(baseURL + "/no-group")
	require.NoError(t, err)
	assert.False(t, groupMiddlewareCalled, "Group middleware should not be called")
}

func TestMiddleware_SingleRoute(t *testing.T) {
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
}

func TestMiddleware_StandardResponse(t *testing.T) {
	teardown := setupServer(t, func(s *mhttp.Server) {
		s.Use(mhttp.MiddlewareResponse())
		s.Bind(&TestMiddlewareController{})
	})
	defer teardown()

	t.Run("Success Case", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/success")
		require.NoError(t, err, "Should get response without error")
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "Should read body without error")

		var respMap map[string]interface{}
		err = json.Unmarshal(body, &respMap)
		require.NoError(t, err, "Response should be valid JSON")

		assert.Equal(t, float64(0), respMap["code"])
		assert.Equal(t, "success", respMap["message"])
	})

	t.Run("Error Case", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/error")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		var respMap map[string]interface{}
		err = json.Unmarshal(body, &respMap)
		require.NoError(t, err)

		assert.Equal(t, float64(200), respMap["code"]) // internal error code
		assert.Equal(t, "something went wrong", respMap["message"])
		assert.Nil(t, respMap["data"])
	})
}

func TestMiddleware_Abort(t *testing.T) {
	var handlerCalled bool
	teardown := setupServer(t, func(s *mhttp.Server) {
		s.Use(func(r *mhttp.Request) {
			r.String(http.StatusUnauthorized, "aborted")
			r.Abort()
		})
		s.GET("/test", func(r *mhttp.Request) {
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
}
