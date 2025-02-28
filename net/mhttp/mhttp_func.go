package mhttp

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

func (s *Server) registerDoc(ctx context.Context) {
	s.initOpenAPI(ctx)

	if s.config.OpenapiPath != "" {
		s.BindHandler("GET", s.config.OpenapiPath, s.openapiHandler)
		s.Logger().Infof(ctx, "OpenAPI specification registered at %s", s.config.OpenapiPath)
	}

	if s.config.SwaggerPath != "" {
		s.BindHandler("GET", s.config.SwaggerPath, s.swaggerHandler)
		s.Logger().Infof(ctx, "Swagger UI registered at %s", s.config.SwaggerPath)
	}
}

func (s *Server) doPrintRoute(ctx context.Context) {
	// 打印服务信息
	s.Logger().Infof(ctx, "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)
	// 打印路由信息
	fmt.Printf("%s\n", strings.Repeat("-", 60))
	fmt.Printf("%-10s | %-7s | %-15s \n", "ADDRESS", "METHOD", "ROUTE")
	routes := s.Engine.Routes()
	for _, route := range routes {
		if route.Path == s.config.OpenapiPath || route.Path == s.config.SwaggerPath {
			continue
		}
		fmt.Printf("%-10s | %-7s | %-15s \n",
			s.config.Address,
			route.Method,
			route.Path,
		)
	}
	fmt.Printf("%s\n\n", strings.Repeat("-", 60))
}

// registerValidateTranslator 注册 gin validator 翻译器
func (s *Server) registerValidateTranslator(locale string) {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		zhT := zh.New()
		enT := en.New()
		uni := ut.New(enT, zhT, enT)
		trans, _ := uni.GetTranslator(locale)
		s.translator = trans
		switch locale {
		case "en":
			_ = en_translations.RegisterDefaultTranslations(v, trans)
		case "zh":
			_ = zh_translations.RegisterDefaultTranslations(v, trans)
		}
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			return fld.Tag.Get("dc")
		})
	}
}
