package mhttp

import (
	"context"
	"strings"

	"github.com/graingo/maltose/util/mmeta"
)

func (s *Server) registerDoc(ctx context.Context) {
	s.initOpenAPI(ctx)

	if s.config.OpenapiPath != "" {
		s.GET(s.config.OpenapiPath, s.openapiHandler)
		s.Logger().Infof(ctx, "OpenAPI specification registered at %s", s.config.OpenapiPath)
	}

	if s.config.SwaggerPath != "" {
		s.GET(s.config.SwaggerPath, s.swaggerHandler)
		s.Logger().Infof(ctx, "Swagger UI registered at %s", s.config.SwaggerPath)
	}
}

func (s *Server) initOpenAPI(_ context.Context) {
	if s.config.OpenapiPath == "" {
		return
	}

	spec := &OpenAPI{
		Openapi: "3.0.0",
		Info: Info{
			Title:   s.config.ServerName,
			Version: "1.0.0",
		},
		Paths: make(map[string]PathItem),
	}

	for _, route := range s.Routes() {
		// Only handle controller routes
		if route.Type != routeTypeController {
			continue
		}

		// Use the saved type information directly
		reqType := route.ReqType
		respType := route.RespType

		metaData := mmeta.Data(reqType)
		if len(metaData) == 0 {
			continue
		}

		summary := mmeta.Get(reqType, "summary").String()
		tags := mmeta.Get(reqType, "tag").String()
		description := mmeta.Get(reqType, "dc").String()

		operation := &Operation{
			Tags:        []string{tags},
			Summary:     summary,
			Description: description,
			Responses: map[string]Response{
				"200": {
					Description: "Success",
					Content: map[string]MediaType{
						"application/json": {
							Schema: createResponseSchema(respType),
						},
					},
				},
			},
		}

		if route.Method == "GET" || route.Method == "DELETE" {
			operation.Parameters = createParameters(reqType)
		} else {
			operation.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: createRequestSchema(reqType),
					},
				},
			}
		}

		pathItem := spec.Paths[route.Path]
		switch strings.ToUpper(route.Method) {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		case "HEAD":
			pathItem.Head = operation
		}
		spec.Paths[route.Path] = pathItem
	}

	s.openapi = spec
}

// openapiHandler handles OpenAPI requests.
func (s *Server) openapiHandler(r *Request) {
	if s.openapi == nil {
		r.String(500, "OpenAPI specification is not properly initialized")
		return
	}
	r.JSON(200, s.openapi)
}

// swaggerHandler handles Swagger requests.
func (s *Server) swaggerHandler(r *Request) {
	template := defaultSwaggerTemplate
	if s.config.SwaggerTemplate != "" {
		template = s.config.SwaggerTemplate
	}
	r.Header("Content-Type", "text/html")
	if s.config.OpenapiPath == "" {
		r.String(200, "swagger path is empty")
		r.Abort()
	}
	r.String(200, template, s.config.OpenapiPath)
}
