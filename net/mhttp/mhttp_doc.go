package mhttp

import (
	"context"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/graingo/maltose/net/mhttp/contract"
)

// LoadOpenAPI loads a generated document and its manifest before Prepare.
// The bytes are copied so callers may release or reuse their buffers.
func (s *Server) LoadOpenAPI(document, manifest []byte) error {
	loader := openapi3.NewLoader()
	parsed, err := loader.LoadFromData(document)
	if err != nil {
		return fmt.Errorf("load OpenAPI: %w", err)
	}
	if err = parsed.Validate(context.Background()); err != nil {
		return fmt.Errorf("validate OpenAPI: %w", err)
	}
	s.openapi = append([]byte(nil), document...)
	s.openapiManifest = append([]byte(nil), manifest...)
	return nil
}

// LoadOpenAPIFile reads filename and filename + ".manifest.json".
func (s *Server) LoadOpenAPIFile(filename string) error {
	document, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	manifest, err := os.ReadFile(filename + ".manifest.json")
	if err != nil {
		return err
	}
	return s.LoadOpenAPI(document, manifest)
}
func (s *Server) ValidateOpenAPIRoutes() error {
	if len(s.openapi) == 0 {
		return fmt.Errorf("load generated OpenAPI and manifest before enabling documentation")
	}
	operations := s.contracts()
	return contract.CheckArtifact(s.openapi, s.openapiManifest, operations)
}
func (s *Server) contracts() []*contract.Operation {
	operations := []*contract.Operation{}
	for _, route := range s.routes {
		if route.contract != nil {
			operations = append(operations, route.contract)
		}
	}
	return operations
}
func (s *Server) registerDoc(_ context.Context) {
	if s.config.OpenapiPath != "" {
		s.GET(s.config.OpenapiPath, s.openapiHandler)
	}
	if s.config.SwaggerPath != "" {
		s.GET(s.config.SwaggerPath, s.swaggerHandler)
	}
}
func (s *Server) openapiHandler(r *Request) {
	media := "application/yaml"
	if len(s.openapi) > 0 && s.openapi[0] == '{' {
		media = "application/json; charset=utf-8"
	}
	r.Data(200, media, s.openapi)
}
func (s *Server) swaggerHandler(r *Request) {
	template := defaultSwaggerTemplate
	if s.config.SwaggerTemplate != "" {
		template = s.config.SwaggerTemplate
	}
	r.Header("Content-Type", "text/html")
	r.String(200, template, s.config.OpenapiPath)
}
