package mhttp_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Controller for Validation ---

type TestValidationController struct{}

type ValidationReq struct {
	mmeta.Meta `path:"/validate" method:"post"`
	Name       string `json:"name" form:"name" binding:"required,min=2,max=10"`
	Age        int    `json:"age" form:"age" binding:"required,min=1,max=120"`
	Email      string `json:"email" form:"email" binding:"required,email"`
}
type ValidationRes struct {
	Message string `json:"message"`
}

func (c *TestValidationController) DoValidation(ctx context.Context, req *ValidationReq) (*ValidationRes, error) {
	return &ValidationRes{Message: "valid"}, nil
}

type CustomValidationReq struct {
	mmeta.Meta `path:"/custom-validate" method:"post"`
	// The value must be "maltose"
	Framework string `json:"framework" binding:"required,is-maltose"`
}
type CustomValidationRes struct {
	Message string `json:"message"`
}

func (c *TestValidationController) DoCustomValidation(ctx context.Context, req *CustomValidationReq) (*CustomValidationRes, error) {
	return &CustomValidationRes{Message: "valid"}, nil
}

// --- Tests ---

func TestValidate_BuiltIn(t *testing.T) {
	testCases := []struct {
		name         string
		payload      string
		expectedMsg  string
		expectedCode int
	}{
		{
			name:         "Missing Name",
			payload:      `{"age":18, "email":"test@example.com"}`,
			expectedMsg:  "name为必填字段",
			expectedCode: 400,
		},
		{
			name:         "Name too short",
			payload:      `{"name":"a", "age":18, "email":"test@example.com"}`,
			expectedMsg:  "name长度必须至少为2个字符",
			expectedCode: 400,
		},
		{
			name:         "Age too small",
			payload:      `{"name":"test", "age":0, "email":"test@example.com"}`,
			expectedMsg:  "age为必填字段",
			expectedCode: 400,
		},
		{
			name:         "Invalid Email",
			payload:      `{"name":"test", "age":18, "email":"invalid-email"}`,
			expectedMsg:  "email必须是一个有效的邮箱",
			expectedCode: 400,
		},
		{
			name:         "Valid",
			payload:      `{"name":"test", "age":18, "email":"test@example.com"}`,
			expectedMsg:  `{"code":0,"message":"OK","data":{"message":"valid"}}`,
			expectedCode: 200,
		},
	}

	teardown := setupServer(t, func(s *mhttp.Server) {
		s.Use(mhttp.MiddlewareResponse())
		s.Bind(&TestValidationController{})
	})
	defer teardown()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bodyReader := strings.NewReader(tc.payload)
			resp, err := http.Post(baseURL+"/validate", "application/json", bodyReader)
			require.NoError(t, err)
			defer resp.Body.Close()

			body, _ := ioutil.ReadAll(resp.Body)
			assert.Equal(t, tc.expectedCode, resp.StatusCode)

			if tc.expectedCode == 200 {
				assert.JSONEq(t, tc.expectedMsg, string(body))
			} else {
				assert.Contains(t, string(body), tc.expectedMsg)
			}
		})
	}
}

func TestValidate_CustomRule(t *testing.T) {
	teardown := setupServer(t, func(s *mhttp.Server) {
		// Define custom rule
		isMaltoseRule := func(fl validator.FieldLevel) bool {
			return fl.Field().String() == "maltose"
		}
		// Define translation
		translations := map[string]string{
			"zh": "{0}必须是maltose",
			"en": "{0} must be maltose",
		}
		s.RegisterRuleWithTranslation("is-maltose", isMaltoseRule, translations)
		s.Use(mhttp.MiddlewareResponse())
		s.Bind(&TestValidationController{})
	})
	defer teardown()

	// Test case that fails custom validation
	bodyReader := strings.NewReader(`{"framework":"other"}`)
	resp, err := http.Post(baseURL+"/custom-validate", "application/json", bodyReader)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "framework必须是maltose")

	// Test case that passes custom validation
	bodyReader = strings.NewReader(`{"framework":"maltose"}`)
	resp, err = http.Post(baseURL+"/custom-validate", "application/json", bodyReader)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"code":0, "message":"OK", "data":{"message":"valid"}}`, string(body))
}
