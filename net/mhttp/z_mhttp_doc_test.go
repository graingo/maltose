package mhttp_test

import (
	"context"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/require"
)

type TestDocController struct{}
type DocReq struct {
	mmeta.Meta `path:"/doc/test" method:"POST" operation_id:"createDoc" status:"201"`
	ID         int `json:"id" field:"required" binding:"min=0"`
}
type DocRes struct {
	Status string `json:"status"`
}

func (*TestDocController) Create(_ context.Context, _ *DocReq) (*DocRes, error) {
	return &DocRes{"created"}, nil
}
func TestDocumentation(t *testing.T) {
	op, err := contract.Compile(contract.TypeOf[DocReq](), contract.TypeOf[DocRes]())
	require.NoError(t, err)
	for _, version := range []string{"3.0.0", "3.1.0"} {
		t.Run(version, func(t *testing.T) {
			artifact, err := contract.Generate([]*contract.Operation{op}, contract.Options{Version: version, Format: "json"}, func(e *contract.Extensions) error {
				if err := e.Info(contract.Info{Title: "Docs API", Version: "0.2.0", Contact: &contract.Contact{Name: "Owner"}}); err != nil {
					return err
				}
				if err := e.Server(contract.Server{URL: "/"}); err != nil {
					return err
				}
				return e.Tag(contract.Tag{Name: "Docs", Description: "Document operations"})
			})
			require.NoError(t, err)
			for _, file := range []bool{false, true} {
				s := mhttp.New()
				s.SetConfigWithMap(map[string]any{"openapi_path": "/api.json", "swagger_path": "/swagger"})
				s.Bind(&TestDocController{})
				if file {
					name := filepath.Join(t.TempDir(), "openapi.json")
					require.NoError(t, os.WriteFile(name, artifact.Document, 0600))
					require.NoError(t, os.WriteFile(name+".manifest.json", artifact.Manifest, 0600))
					require.NoError(t, s.LoadOpenAPIFile(name))
				} else {
					require.NoError(t, s.LoadOpenAPI(artifact.Document, artifact.Manifest))
				}
				require.NoError(t, s.Prepare(context.Background()))
				w := httptest.NewRecorder()
				s.ServeHTTP(w, httptest.NewRequest("GET", "/api.json", nil))
				require.Equal(t, 200, w.Code)
				require.Equal(t, artifact.Document, w.Body.Bytes())
				w = httptest.NewRecorder()
				s.ServeHTTP(w, httptest.NewRequest("GET", "/swagger", nil))
				require.Contains(t, w.Body.String(), `url: "/api.json"`)
				req := httptest.NewRequest("POST", "/doc/test", strings.NewReader(`{"id":0}`))
				req.Header.Set("Content-Type", "application/json")
				w = httptest.NewRecorder()
				s.ServeHTTP(w, req)
				require.Equal(t, 201, w.Code)
				require.JSONEq(t, `{"status":"created"}`, w.Body.String())
			}
		})
	}
	s := mhttp.New()
	s.SetConfigWithMap(map[string]any{"openapi_path": "/api.json"})
	require.ErrorContains(t, s.Prepare(context.Background()), "load generated")
}
