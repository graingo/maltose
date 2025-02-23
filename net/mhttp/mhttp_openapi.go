package mhttp

import (
	"encoding/json"
	"strings"
)

// OpenAPIInfo OpenAPI 基本信息
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// GenerateOpenAPI 生成 OpenAPI 文档
func (s *Server) GenerateOpenAPI(info OpenAPIInfo) ([]byte, error) {
	docs := map[string]interface{}{
		"openapi": "3.0.0",
		"info":    info,
		"paths":   make(map[string]interface{}),
	}

	paths := make(map[string]interface{})
	for _, meta := range s.metadata {
		pathItem := map[string]interface{}{
			strings.ToLower(meta.Method): map[string]interface{}{
				"summary":     meta.Summary,
				"description": meta.Description,
				"tags":        meta.Tags,
				"requestBody": map[string]interface{}{
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": meta.Request,
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Success",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": meta.Response,
							},
						},
					},
				},
			},
		}
		paths[meta.Path] = pathItem
	}
	docs["paths"] = paths

	return json.Marshal(docs)
}
