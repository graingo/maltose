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
	config     ServerConfig
	middleware []gin.HandlerFunc
	openapi    *OpenAPI
}

// New 创建新的 HTTP 服务实例
func New() *Server {
	// 禁用 gin 的默认日志输出
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	// 设置为生产模式
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		Engine: gin.New(),
		config: NewConfig(),
	}
	// 添加默认中间件
	s.Use(internalMiddlewareServerTracing())

	return s
}

// Run 启动 HTTP 服务
func (s *Server) Run() {
	ctx := context.Background()
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
