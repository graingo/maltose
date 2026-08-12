package mhttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfigMutationAndCloning(t *testing.T) {
	defaults := cloneConfig(nil)
	assert.Equal(t, defaultPort, defaults.Address)
	assert.NotNil(t, defaults.Logger)

	original := &Config{Address: ":9000"}
	cloned := cloneConfig(original)
	assert.NotSame(t, original, cloned)
	assert.NotNil(t, cloned.Logger)

	require.NoError(t, cloned.SetConfigWithMap(map[string]any{
		"server_name": "mapped", "read_timeout": "2s", "max_header_bytes": 2048,
	}))
	assert.Equal(t, "mapped", cloned.ServerName)
	assert.Equal(t, 2*time.Second, cloned.ReadTimeout)
	assert.Equal(t, 2048, cloned.MaxHeaderBytes)

	config, err := ConfigFromMap(map[string]any{"address": ":8088", "server_locale": ""})
	require.NoError(t, err)
	server := New(config)
	server.SetAddress(":8089")
	server.SetServerName("api")
	server.SetLogger(nil)
	assert.Equal(t, ":8089", server.config.Address)
	assert.Equal(t, "api", server.config.ServerName)
	assert.NotNil(t, server.logger())

	require.NoError(t, server.SetConfigWithMap(map[string]any{"health_check": "/ready"}))
	assert.Equal(t, "/ready", server.config.HealthCheck)
	server.SetConfig(&Config{Address: ":8090"})
	assert.Equal(t, ":8090", server.config.Address)
	assert.NotNil(t, server.config.Logger)

	_, err = ConfigFromMap(map[string]any{"max_header_bytes": []int{1}})
	assert.Error(t, err)
	server.SetLogger(mlog.New())
	assert.NotNil(t, server.logger())
}

func TestRouterMethodHelpersAndNoRouteHandler(t *testing.T) {
	server := New(&Config{ServerLocale: ""})
	handler := func(*Request) {}
	server.POST("/post", handler)
	server.PUT("/put", handler)
	server.DELETE("/delete", handler)
	server.HEAD("/head", handler)
	server.OPTIONS("/options", handler)
	server.PATCH("/patch", handler)
	server.Any("/any", handler)
	assert.Len(t, server.Routes(), 13)

	server.SetNoRouteHandler(func(request *Request) {
		request.String(http.StatusTeapot, "missing")
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	server.engine.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusTeapot, recorder.Code)
	assert.Equal(t, "missing", recorder.Body.String())
}

func TestRequestAccessors(t *testing.T) {
	server := New(&Config{ServerName: "request-test", ServerLocale: ""})
	recorder := httptest.NewRecorder()
	ginContext, engine := gin.CreateTestContext(recorder)
	assert.NotNil(t, engine)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	request := newRequest(ginContext, server)

	assert.Same(t, request, RequestFromCtx(request.Request.Context()))
	assert.Equal(t, "request-test", request.GetServerName())
	assert.Same(t, server.config, request.Conf())
	assert.Same(t, server.logger(), request.Logger())
	request.SetHandlerResponse("response")
	assert.Equal(t, "response", request.GetHandlerResponse())
	err := errors.New("request error")
	assert.Same(t, request, request.Error(err))
	require.Len(t, request.Errors, 1)
	assert.ErrorIs(t, request.Errors[0].Err, err)
	assert.Nil(t, request.GetTranslator())
	assert.Nil(t, RequestFromCtx(context.WithValue(context.Background(), testRequestContextKey{}, request)))
}

type testRequestContextKey struct{}
