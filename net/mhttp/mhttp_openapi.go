package mhttp

import (
	"reflect"
	"strings"
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
