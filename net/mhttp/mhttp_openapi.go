package mhttp

import (
	"context"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// OpenAPI OpenAPI 规范对象
type OpenAPI struct {
	*openapi3.T
}

// initOpenAPI 初始化 OpenAPI 规范
func (s *Server) initOpenAPI() {
	if s.config.OpenapiPath == "" {
		return
	}

	ctx := context.Background()
	swagger := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   s.config.ServerName,
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{}, // 使用 make 初始化 map
	}

	// 遍历所有路由，生成 OpenAPI 文档
	for _, route := range s.Routes() {
		path := route.Path
		method := route.Method

		// 跳过中间件和内部路由
		if path == s.config.OpenapiPath || path == s.config.SwaggerPath {
			continue
		}

		// 创建响应对象
		// successResponse := openapi3.NewResponse().
		// 	WithDescription("Success").
		// 	WithContent(openapi3.NewContentWithJSONSchema(
		// 		openapi3.NewSchema().
		// 			WithProperty("code", openapi3.NewIntegerSchema()).
		// 			WithProperty("message", openapi3.NewStringSchema()).
		// 			WithProperty("data", openapi3.NewObjectSchema()),
		// 	))
		// 返回对象数组
		responses := openapi3.NewResponses()
		responses.Default().Value.WithContent(openapi3.NewContentWithJSONSchema(
			openapi3.NewSchema().
				WithProperty("code", openapi3.NewIntegerSchema()).
				WithProperty("message", openapi3.NewStringSchema()).
				WithProperty("data", openapi3.NewObjectSchema())))
		// responses.Set("200", successResponse)

		operation := &openapi3.Operation{
			Summary:   route.Path,
			Tags:      []string{s.config.ServerName},
			Responses: responses,
		}

		// 确保路径存在

		// 根据 HTTP 方法设置操作
		pathItem := &openapi3.PathItem{}
		switch method {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		}

		swagger.Paths.Set(path, pathItem)
	}

	// 验证 OpenAPI 文档
	if err := swagger.Validate(ctx); err != nil {
		s.Logger().Errorf(ctx, "OpenAPI validation failed: %v", err)
		return
	}

	s.openapi = &OpenAPI{swagger}
}

// openapiHandler OpenAPI 文档处理器
func (s *Server) openapiHandler(r *Request) {
	if s.openapi == nil {
		r.String(400, "OpenAPI specification is not available")
		return
	}
	r.JSON(200, s.openapi.T)
}

func (s *Server) Registered() {
	// 初始化 OpenAPI
	s.initOpenAPI()

	// 注册 OpenAPI 和 Swagger UI 路由
	if s.config.OpenapiPath != "" {
		s.GET(s.config.OpenapiPath, func(c *gin.Context) {
			s.openapiHandler(newRequest(c, s))
		})
	}

	if s.config.SwaggerPath != "" {
		s.GET(s.config.SwaggerPath, func(c *gin.Context) {
			s.swaggerHandler(newRequest(c, s))
		})
	}
}
