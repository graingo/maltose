package mhttp_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Controllers ---

type TestRouterController struct{}

type HelloReq struct {
	mmeta.Meta `path:"/hello" method:"get"`
	Name       string `json:"name" form:"name"`
}
type HelloRes struct {
	Message string `json:"message"`
}

func (c *TestRouterController) Hello(_ context.Context, req *HelloReq) (*HelloRes, error) {
	return &HelloRes{Message: fmt.Sprintf("Hello, %s", req.Name)}, nil
}

type UserReq struct {
	mmeta.Meta `path:"/user/:id" method:"post"`
	ID         int    `json:"-" uri:"id"`
	Content    string `json:"content"`
}
type UserRes struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

func (c *TestRouterController) UpdateUser(_ context.Context, req *UserReq) (*UserRes, error) {
	return &UserRes{ID: req.ID, Content: req.Content}, nil
}

// --- Tests ---

func TestRouter(t *testing.T) {
	t.Run("http_methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		teardown := setupServer(t, func(s *mhttp.Server) {
			for _, method := range methods {
				path := "/" + strings.ToLower(method)
				s.Handle(method, path, func(r *mhttp.Request) {
					r.String(http.StatusOK, "ok")
				})
			}
		})
		defer teardown()

		for _, method := range methods {
			t.Run(strings.ToLower(method), func(t *testing.T) {
				req, err := http.NewRequest(method, baseURL+"/"+strings.ToLower(method), nil)
				require.NoError(t, err)

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
		}
	})

	t.Run("grouping", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			v1 := s.Group("/v1")
			v1.GET("/ping", func(r *mhttp.Request) {
				r.String(http.StatusOK, "v1 pong")
			})

			v2 := s.Group("/v2")
			v2.GET("/ping", func(r *mhttp.Request) {
				r.String(http.StatusOK, "v2 pong")
			})
		})
		defer teardown()

		// Test V1
		respV1, errV1 := http.Get(baseURL + "/v1/ping")
		require.NoError(t, errV1)
		defer respV1.Body.Close()
		bodyV1, _ := ioutil.ReadAll(respV1.Body)
		assert.Equal(t, "v1 pong", string(bodyV1))

		// Test V2
		respV2, errV2 := http.Get(baseURL + "/v2/ping")
		require.NoError(t, errV2)
		defer respV2.Body.Close()
		bodyV2, _ := ioutil.ReadAll(respV2.Body)
		assert.Equal(t, "v2 pong", string(bodyV2))
	})

	t.Run("controller_binding_simple_get", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.Bind(&TestRouterController{})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/hello?name=Maltose")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(t, `{"message":"Hello, Maltose"}`, string(body))
	})

	t.Run("controller_binding_path_and_body", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.Bind(&TestRouterController{})
		})
		defer teardown()

		bodyReader := strings.NewReader(`{"content":"new content"}`)
		resp, err := http.Post(baseURL+"/user/123", "application/json", bodyReader)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		assert.JSONEq(t, `{"id":123, "content":"new content"}`, string(body))
	})

	t.Run("uri_param_extraction", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.GET("/param/:name", func(r *mhttp.Request) {
				name := r.Param("name")
				r.String(http.StatusOK, name)
			})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/param/world")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, "world", string(body))
	})
}
