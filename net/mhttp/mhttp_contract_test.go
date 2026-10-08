package mhttp_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/require"
)

type ProtectedReq struct {
	mmeta.Meta `method:"GET" group:"/api" path:"/products/:id" operation_id:"product.get"`
	ID         string `path:"id" binding:"uuid"`
}
type ProtectedRes struct {
	ID string `json:"id"`
}
type LogoutReq struct {
	mmeta.Meta `method:"POST" group:"/api" path:"/logout" status:"204"`
	Reason     string `json:"reason" binding:"min=2"`
}
type LogoutRes struct{}
type contractController struct{}

func (*contractController) Get(_ context.Context, r *ProtectedReq) (*ProtectedRes, error) {
	return &ProtectedRes{r.ID}, nil
}
func (*contractController) Logout(_ context.Context, _ *LogoutReq) (*LogoutRes, error) {
	return nil, nil
}
func TestAuthenticationBeforeBindingAndNoContent(t *testing.T) {
	s := mhttp.New()
	s.Group("/api", func(g *mhttp.RouterGroup) {
		g.Middleware(func(r *mhttp.Request) {
			if r.OperationID() == "product.get" && r.GetHeader("Authorization") == "" {
				r.Status(401)
				r.Writer.WriteHeaderNow()
				r.Abort()
				return
			}
			r.Next()
		})
		g.Bind(&contractController{})
	})
	require.NoError(t, s.Prepare(context.Background()))
	for _, auth := range []string{"", "session"} {
		r := httptest.NewRequest("GET", "/api/products/not-a-uuid", nil)
		r.Header.Set("Authorization", auth)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if auth == "" {
			require.Equal(t, 401, w.Code)
		} else {
			require.Equal(t, 400, w.Code)
		}
	}
	r := httptest.NewRequest("POST", "/api/logout", strings.NewReader(`{"reason":"ok"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	require.Equal(t, 204, w.Code)
	require.Empty(t, w.Body.String())
	op, err := contract.Compile(contract.TypeOf[LogoutReq](), contract.TypeOf[LogoutRes]())
	require.NoError(t, err)
	a, err := contract.Generate([]*contract.Operation{op}, contract.Options{Format: "json"}, nil)
	require.NoError(t, err)
	require.Contains(t, string(a.Document), `"requestBody"`)
	require.NotContains(t, string(a.Document), `"description": "OK"`)
}

type FormReq struct {
	mmeta.Meta `method:"POST" path:"/form"`
	Name       string `form:"name" field:"required"`
	Query      string `query:"name"`
	Header     string `header:"X-Name"`
}
type FormRes struct {
	Name   string `json:"name"`
	Query  string `json:"query"`
	Header string `json:"header"`
}
type formController struct{}

func (*formController) Form(_ context.Context, r *FormReq) (*FormRes, error) {
	return &FormRes{r.Name, r.Query, r.Header}, nil
}
func TestIndependentSources(t *testing.T) {
	s := mhttp.New()
	s.Bind(&formController{})
	require.NoError(t, s.Prepare(context.Background()))
	r := httptest.NewRequest("POST", "/form?name=query", strings.NewReader("name=form"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("X-Name", "header")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	require.Equal(t, 200, w.Code)
	require.JSONEq(t, `{"name":"form","query":"query","header":"header"}`, w.Body.String())
}
