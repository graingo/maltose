package mhttp

import (
	"context"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/graingo/maltose/util/mmeta"
)

// schemaBuilder is a helper for building OpenAPI schemas at runtime.
type schemaBuilder struct {
	spec *openapi3.T
}

func (s *Server) registerDoc(ctx context.Context) {
	s.initOpenAPI(ctx)

	if s.config.OpenapiPath != "" {
		s.GET(s.config.OpenapiPath, s.openapiHandler)
		s.logger().Infof(ctx, "OpenAPI specification registered at %s", s.config.OpenapiPath)
	}

	if s.config.SwaggerPath != "" {
		s.GET(s.config.SwaggerPath, s.swaggerHandler)
		s.logger().Infof(ctx, "Swagger UI registered at %s", s.config.SwaggerPath)
	}
}

func (s *Server) initOpenAPI(_ context.Context) {
	if s.config.OpenapiPath == "" {
		return
	}

	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   s.config.ServerName,
			Version: "1.0.0",
		},
		Paths:      openapi3.NewPaths(),
		Components: &openapi3.Components{},
	}
	spec.Components.Schemas = make(openapi3.Schemas)

	builder := &schemaBuilder{spec: spec}

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

		operation := &openapi3.Operation{
			Tags:        []string{tags},
			Summary:     summary,
			Description: description,
			Responses:   openapi3.NewResponses(),
		}
		operation.Responses.Set("200", &openapi3.ResponseRef{
			Value: openapi3.NewResponse().
				WithDescription("Success").
				WithContent(openapi3.NewContentWithJSONSchema(builder.createSchema(respType))),
		})

		if route.Method == "GET" || route.Method == "DELETE" {
			operation.Parameters = builder.createParameters(reqType)
		} else {
			operation.RequestBody = &openapi3.RequestBodyRef{
				Value: openapi3.NewRequestBody().
					WithRequired(true).
					WithContent(openapi3.NewContentWithJSONSchema(builder.createSchema(reqType))),
			}
		}

		pathItem := spec.Paths.Find(route.Path)
		if pathItem == nil {
			pathItem = &openapi3.PathItem{}
		}

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
		spec.Paths.Set(route.Path, pathItem)
	}

	s.openapi = spec
}

func (b *schemaBuilder) createParameters(reqType reflect.Type) openapi3.Parameters {
	params := openapi3.NewParameters()
	if reqType.Kind() != reflect.Struct {
		return params
	}
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		if field.Anonymous { // Skip embedded structs like m.Meta
			continue
		}

		schema := b.typeToSchema(field.Type)
		if schema.Value == nil {
			continue
		}

		paramName := field.Tag.Get("form")
		if paramName == "" {
			paramName = field.Name
		}

		param := openapi3.NewQueryParameter(paramName).
			WithSchema(schema.Value).
			WithDescription(field.Tag.Get("dc"))

		if strings.Contains(field.Tag.Get("binding"), "required") {
			param.Required = true
		}

		params = append(params, &openapi3.ParameterRef{Value: param})
	}
	return params
}

func (b *schemaBuilder) createSchema(p reflect.Type) *openapi3.Schema {
	return b.typeToSchema(p).Value
}

func (b *schemaBuilder) typeToSchema(p reflect.Type) *openapi3.SchemaRef {
	if p == nil {
		return &openapi3.SchemaRef{Value: openapi3.NewObjectSchema()}
	}
	// Handle pointers
	if p.Kind() == reflect.Ptr {
		ref := b.typeToSchema(p.Elem())
		if ref.Value != nil {
			ref.Value.Nullable = true
		}
		return ref
	}

	// Handle slices/arrays
	if p.Kind() == reflect.Slice || p.Kind() == reflect.Array {
		itemsRef := b.typeToSchema(p.Elem())
		schema := openapi3.NewArraySchema()
		schema.Items = itemsRef
		return &openapi3.SchemaRef{Value: schema}
	}
	// Handle primitive types
	switch p.Kind() {
	case reflect.String:
		return &openapi3.SchemaRef{Value: openapi3.NewStringSchema()}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &openapi3.SchemaRef{Value: openapi3.NewIntegerSchema()}
	case reflect.Float32, reflect.Float64:
		return &openapi3.SchemaRef{Value: openapi3.NewFloat64Schema()}
	case reflect.Bool:
		return &openapi3.SchemaRef{Value: openapi3.NewBoolSchema()}
	case reflect.Interface:
		return &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{openapi3.TypeObject}, AdditionalProperties: openapi3.AdditionalProperties{Has: openapi3.BoolPtr(true)}}}
	}
	if p.Kind() == reflect.Struct {
		// Handle custom struct types by creating a component schema
		cleanTypeName := p.Name()

		// If the schema is not already in components, create it.
		if _, ok := b.spec.Components.Schemas[cleanTypeName]; !ok {
			// Add a placeholder to components to prevent infinite recursion for self-referencing structs.
			b.spec.Components.Schemas[cleanTypeName] = &openapi3.SchemaRef{Value: openapi3.NewObjectSchema()}
			// Build the full schema for the struct.
			schema := openapi3.NewObjectSchema()
			for i := 0; i < p.NumField(); i++ {
				field := p.Field(i)
				if field.Anonymous { // Skip embedded structs like m.Meta
					continue
				}

				jsonTag := field.Tag.Get("json")
				if jsonTag == "" || jsonTag == "-" {
					continue
				}
				jsonName := strings.Split(jsonTag, ",")[0]

				fieldSchemaRef := b.typeToSchema(field.Type)
				if field.Tag.Get("dc") != "" && fieldSchemaRef.Value != nil {
					fieldSchemaRef.Value.Description = field.Tag.Get("dc")
				}
				schema.Properties[jsonName] = fieldSchemaRef
			}

			// Replace the placeholder with the fully constructed schema.
			b.spec.Components.Schemas[cleanTypeName] = &openapi3.SchemaRef{Value: schema}
		}
		// Return a reference to the component schema.
		return openapi3.NewSchemaRef("#/components/schemas/"+cleanTypeName, nil)
	}

	return &openapi3.SchemaRef{Value: openapi3.NewObjectSchema()}
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
