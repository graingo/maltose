package mclient_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/graingo/maltose/net/mclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingResponseBody struct{}

func (failingResponseBody) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (failingResponseBody) Close() error             { return nil }

func TestResponseMethodsHandleNilReceiver(t *testing.T) {
	var response *mclient.Response

	assert.Empty(t, response.GetCookie("session"))
	assert.Nil(t, response.GetCookies())
	assert.Nil(t, response.GetCookieMap())
	assert.Empty(t, response.ReadAll())
	assert.Empty(t, response.ReadAllString())
	assert.False(t, response.IsSuccess())
	assert.Error(t, response.Parse(new(string)))
	assert.NoError(t, response.Close())
	assert.Nil(t, response.GetResult())
	assert.Nil(t, response.GetError())
	assert.NotPanics(t, func() {
		response.SetBodyContent([]byte("body"))
		response.SetResult("result")
		response.SetError("error")
	})
}

func TestResponseMethodsHandleMissingBodyAndRequest(t *testing.T) {
	response := &mclient.Response{Response: &http.Response{}}
	assert.Empty(t, response.ReadAll())
	assert.NoError(t, response.Close())
	assert.Error(t, response.Parse(new(string)))

	response.Body = failingResponseBody{}
	assert.NotPanics(t, func() {
		assert.Empty(t, response.ReadAll())
	})

	response.SetBodyContent([]byte("restored"))
	require.NotNil(t, response.Body)
	assert.Equal(t, "restored", response.ReadAllString())
}
