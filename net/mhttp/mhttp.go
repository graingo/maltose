package mhttp

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultServerName  = "default"
	defaultPort        = "8080"
	defaultOpenapiPath = "/api.json"
	defaultSwaggerPath = "/swagger"
)

// Server HTTP 服务结构
type Server struct {
	*gin.Engine
	config       ServerConfig
	middlewares  []MiddlewareFunc
	routes       []Route
	openapi      *OpenAPI
	preBindItems []preBindItem
}

func New() *Server {
	// 禁用 gin 的默认日志输出
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	// 设置为生产模式
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		Engine:       gin.New(),
		config:       NewConfig(),
		preBindItems: make([]preBindItem, 0),
	}
	// 添加默认中间件
	s.Use(internalMiddlewareServerTracing(), internalMiddlewareDefaultResponse(), internalMiddlewareRecovery())

	return s
}

// Run 启动 HTTP 服务
func (s *Server) Run() {
	ctx := context.Background()

	// 注册 OpenAPI 和 Swagger
	s.registerDoc(ctx)

	// 在启动前注册所有路由
	s.bindRoutes(ctx)

	// 打印路由信息
	s.doPrintRoute(ctx)

	srv := &http.Server{
		Addr:           s.config.Address,
		Handler:        s.Engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}

	// 创建错误通道
	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// 监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		s.Logger().Errorf(ctx, "HTTP server %s start failed: %v", s.config.ServerName, err)
	case <-quit:
		s.Logger().Infof(ctx, "Shutting down server...")
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			s.Logger().Errorf(ctx, "Server forced to shutdown: %v", err)
		}
	}
}

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
