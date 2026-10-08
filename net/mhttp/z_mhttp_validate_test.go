package mhttp_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/net/mhttp/contract"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/require"
)

type TestValidationController struct{}
type ValidationReq struct {
	mmeta.Meta `path:"/validate" method:"POST"`
	Name       string `json:"name" binding:"required,min=2,max=10"`
	Age        int    `json:"age" binding:"gte=1,lte=120"`
	Email      string `json:"email" binding:"required,email"`
}
type ValidationRes struct {
	Message string `json:"message"`
}

func (*TestValidationController) Validate(_ context.Context, r *ValidationReq) (*ValidationRes, error) {
	return &ValidationRes{"Welcome, " + r.Name}, nil
}
func TestValidation(t *testing.T) {
	s := mhttp.New()
	s.Use(func(r *mhttp.Request) {
		r.Next()
		if len(r.Errors) > 0 {
			var err *contract.ValidationError
			if errors.As(r.Errors.Last().Err, &err) {
				r.JSON(422, err)
				return
			}
		}
	})
	s.Bind(&TestValidationController{})
	require.NoError(t, s.Prepare(context.Background()))
	for _, tc := range []struct{ body, field, code string }{
		{`{"email":"a@b.com"}`, "json.name", "required"},
		{`{"name":"a","email":"a@b.com"}`, "json.name", "min"},
		{`{"name":"test","age":0,"email":"a@b.com"}`, "json.age", "gte"},
		{`{"name":"test","email":"invalid"}`, "json.email", "email"},
		{`{"name":"test","email":"a@b.com"}`, "", ""},
	} {
		t.Run(tc.code, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/validate", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			s.ServeHTTP(w, req)
			if tc.code == "" {
				require.Equal(t, 200, w.Code)
				require.JSONEq(t, `{"message":"Welcome, test"}`, w.Body.String())
			} else {
				require.Equal(t, 422, w.Code)
				require.Contains(t, w.Body.String(), tc.field)
				require.Contains(t, w.Body.String(), tc.code)
			}
		})
	}
}
