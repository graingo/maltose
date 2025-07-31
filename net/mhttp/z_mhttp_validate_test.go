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
	mmeta.Meta `path:"/validate" method:"post" dc:"A route for testing built-in validations."`
	Name       string `json:"name" form:"name" binding:"required,min=2,max=10" dc:"User's name"`
	Age        int    `json:"age" form:"age" binding:"gte=1,lte=120" dc:"User's age"`
	Email      string `json:"email" form:"email" binding:"required,email" dc:"User's email"`
}
type ValidationRes struct {
	Message string `json:"message"`
}

func (c *TestValidationController) DoValidation(_ context.Context, req *ValidationReq) (*ValidationRes, error) {
	return &ValidationRes{Message: "Welcome, " + req.Name}, nil
}

type CustomValidationReq struct {
	mmeta.Meta `path:"/custom-validate" method:"post"`
	// The value must be "maltose"
	Framework string `json:"framework" binding:"required,is-maltose"`
}
type CustomValidationRes struct {
	Message string `json:"message"`
}

func (c *TestValidationController) DoCustomValidation(_ context.Context, _ *CustomValidationReq) (*CustomValidationRes, error) {
	return &CustomValidationRes{Message: "valid"}, nil
}

// --- Tests ---

func TestValidation(t *testing.T) {
	t.Run("built_in_rules", func(t *testing.T) {
		testCases := []struct {
			name         string
			payload      string
			expectedMsg  string
			expectedCode int
		}{
			{
				name:         "missing_name",
				payload:      `{"age":18, "email":"test@example.com"}`,
				expectedMsg:  "User's name为必填字段",
				expectedCode: 400,
			},
			{
				name:         "name_too_short",
				payload:      `{"name":"a", "age":18, "email":"test@example.com"}`,
				expectedMsg:  "User's name长度必须至少为2个字符",
				expectedCode: 400,
			},
			{
				name:         "age_too_small",
				payload:      `{"name":"test", "age":0, "email":"test@example.com"}`,
				expectedMsg:  "User's age必须大于或等于1",
				expectedCode: 400,
			},
			{
				name:         "invalid_email",
				payload:      `{"name":"test", "age":18, "email":"invalid-email"}`,
				expectedMsg:  "User's email必须是一个有效的邮箱",
				expectedCode: 400,
			},
			{
				name:         "valid_request",
				payload:      `{"name":"test", "age":18, "email":"test@example.com"}`,
				expectedMsg:  `{"code":0,"message":"OK","data":{"message":"Welcome, test"}}`,
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
	})

	t.Run("custom_rule", func(t *testing.T) {
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
	})
}
