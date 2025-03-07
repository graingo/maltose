package mhttp

import (
	"context"
	"reflect"
	"strings"

	"github.com/graingo/maltose/util/mmeta"
)

// OpenAPI 规范对象
type OpenAPI struct {
	Openapi string              `json:"openapi"`
	Info    Info                `json:"info"`
	Paths   map[string]PathItem `json:"paths"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type PathItem struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
}

type Operation struct {
	Tags        []string            `json:"tags,omitempty"`
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
	Description string `json:"description,omitempty"`
}

type RequestBody struct {
	Required bool                 `json:"required"`
	Content  map[string]MediaType `json:"content"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type                 string            `json:"type,omitempty"`
	Format               string            `json:"format,omitempty"`
	Properties           map[string]Schema `json:"properties,omitempty"`
	Items                *Schema           `json:"items,omitempty"`
	AdditionalProperties *Schema           `json:"additionalProperties,omitempty"`
	Description          string            `json:"description,omitempty"`
	Required             []string          `json:"required,omitempty"`
}

func (s *Server) openapiHandler(r *Request) {
	if s.openapi == nil {
		r.String(500, "OpenAPI specification is not properly initialized")
		return
	}
	r.JSON(200, s.openapi)
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
		// 只处理控制器路由
		if route.Type != routeTypeController {
			continue
		}

		// 直接使用保存的类型信息
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
		}
		spec.Paths[route.Path] = pathItem
	}

	s.openapi = spec
}

func createParameters(t reflect.Type) []Parameter {
	// 如果是指针类型，获取其基础类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var params []Parameter
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous || field.Name == "Meta" {
			continue
		}

		param := field.Tag.Get("json")
		if param == "" {
			param = field.Tag.Get("form")
		}
		if param == "" || param == "-" {
			continue
		}

		params = append(params, Parameter{
			Name:        param,
			In:          "query",
			Required:    strings.Contains(field.Tag.Get("binding"), "required"),
			Schema:      createSchema(field.Type),
			Description: field.Tag.Get("dc"),
		})
	}
	return params
}

func createSchema(t reflect.Type) Schema {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := Schema{}

	switch t.Kind() {
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int64:
		schema.Type = "integer"
		schema.Format = "int64"
	case reflect.Float64:
		schema.Type = "number"
		schema.Format = "double"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Struct:
		schema.Type = "object"
		schema.Properties = make(map[string]Schema)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Anonymous || field.Name == "Meta" {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			jsonName := strings.Split(jsonTag, ",")[0]

			fieldSchema := createSchema(field.Type)
			fieldSchema.Description = field.Tag.Get("dc")

			if strings.Contains(field.Tag.Get("binding"), "required") {
				schema.Required = append(schema.Required, jsonName)
			}
			schema.Properties[jsonName] = fieldSchema
		}
	case reflect.Slice:
		schema.Type = "array"
		items := createSchema(t.Elem())
		schema.Items = &items
	case reflect.Map:
		schema.Type = "object"
		valueSchema := createSchema(t.Elem())
		schema.AdditionalProperties = &valueSchema
	case reflect.Interface:
		schema.Type = "object"
	}

	return schema
}

func createRequestSchema(t reflect.Type) Schema {
	return createSchema(t)
}

func createResponseSchema(t reflect.Type) Schema {
	return createSchema(t)
}
