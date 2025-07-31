package mhttp_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/graingo/maltose/net/mhttp"
	"github.com/graingo/maltose/util/mmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Test Controller for Documentation ---

type TestDocController struct{}

type DocReq struct {
	mmeta.Meta `path:"/doc/test" method:"post" summary:"Test endpoint" tag:"Documentation" dc:"This is a test endpoint for documentation generation."`
	ID         int `json:"id" binding:"required" dc:"The unique identifier"`
}
type DocRes struct {
	Status string `json:"status" dc:"The status of the operation"`
}

func (c *TestDocController) CreateDoc(_ context.Context, _ *DocReq) (*DocRes, error) {
	return &DocRes{Status: "created"}, nil
}

// --- Tests ---

func TestDocumentation(t *testing.T) {
	t.Run("swagger_ui", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.SetConfigWithMap(map[string]any{
				"openapi_path": "/api/v1/openapi.json",
				"swagger_path": "/api/v1/swagger",
			})
			s.Bind(&TestDocController{})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/api/v1/swagger")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/html", resp.Header.Get("Content-Type"))
		// Check if the HTML contains the correct openapi path
		assert.Contains(t, string(body), `url: "/api/v1/openapi.json"`)
		// Check for a known swagger-ui element
		assert.Contains(t, string(body), `<div id="swagger-ui"></div>`)
	})

	t.Run("openapi_json", func(t *testing.T) {
		teardown := setupServer(t, func(s *mhttp.Server) {
			s.SetConfigWithMap(map[string]any{
				"openapi_path": "/api/v1/openapi.json",
				"swagger_path": "/api/v1/swagger",
			})
			s.Bind(&TestDocController{})
		})
		defer teardown()

		resp, err := http.Get(baseURL + "/api/v1/openapi.json")
		require.NoError(t, err)
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

		// Unmarshal and verify the content
		var openapiSpec map[string]interface{}
		err = json.Unmarshal(body, &openapiSpec)
		require.NoError(t, err, "Should be valid JSON")

		assert.Equal(t, "3.0.0", openapiSpec["openapi"])

		// Check paths
		paths, ok := openapiSpec["paths"].(map[string]interface{})
		require.True(t, ok, "Paths should exist")

		pathItem, ok := paths["/doc/test"].(map[string]interface{})
		require.True(t, ok, "Path /doc/test should exist")

		postOp, ok := pathItem["post"].(map[string]interface{})
		require.True(t, ok, "POST operation should exist")

		// Check operation details
		assert.Equal(t, "Test endpoint", postOp["summary"])
		assert.Contains(t, postOp["tags"], "Documentation")
		assert.Equal(t, "This is a test endpoint for documentation generation.", postOp["description"])

		// Check request body schema by diving into the nested map structure
		reqBody, _ := postOp["requestBody"].(map[string]interface{})
		content, _ := reqBody["content"].(map[string]interface{})
		appJSON, _ := content["application/json"].(map[string]interface{})
		schemaRef, _ := appJSON["schema"].(map[string]interface{})
		ref, _ := schemaRef["$ref"].(string)

		// Verify component schema reference
		assert.Equal(t, "#/components/schemas/DocReq", ref)

		// Check response schema
		responses, _ := postOp["responses"].(map[string]interface{})
		resp200, _ := responses["200"].(map[string]interface{})
		respContent, _ := resp200["content"].(map[string]interface{})
		respAppJSON, _ := respContent["application/json"].(map[string]interface{})
		respSchemaRef, _ := respAppJSON["schema"].(map[string]interface{})
		respRef, _ := respSchemaRef["$ref"].(string)

		assert.Equal(t, "#/components/schemas/DocRes", respRef)
	})
}
