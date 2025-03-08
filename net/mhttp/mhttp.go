package mhttp

import (
	"io"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
)

const (
	DefaultServerName  = "default"
	defaultPort        = "8080"
	defaultOpenapiPath = "/api.json"
	defaultSwaggerPath = "/swagger"
)

// Server HTTP 服务结构
type Server struct {
	RouterGroup
	engine       *gin.Engine
	config       ServerConfig
	routes       []Route
	openapi      *OpenAPI
	preBindItems []preBindItem
	translator   ut.Translator
}

func New() *Server {
	// 禁用 gin 的默认日志输出
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	// 设置为生产模式
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	s := &Server{
		engine:       engine,
		config:       NewConfig(),
		preBindItems: make([]preBindItem, 0),
	}

	// 初始化根 RouterGroup
	s.RouterGroup = RouterGroup{
		server:   s,
		path:     "/",
		ginGroup: engine.Group("/"),
	}

	// 添加默认中间件
	s.Use(
		internalMiddlewareRecovery(),
		internalMiddlewareServerTrace(),
		internalMiddlewareMetric(),
		internalMiddlewareDefaultResponse(),
	)
	if s.config.ServerLocale != "" {
		// 注册翻译器
		s.registerValidateTranslator(s.config.ServerLocale)
	}

	return s
}
