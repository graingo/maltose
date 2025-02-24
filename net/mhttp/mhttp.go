package mhttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DefaultServerName = "default"
	defaultPort       = "8080"
)

// Server HTTP 服务结构
type Server struct {
	*gin.Engine
	config     ServerConfig
	metadata   map[string]*HandlerMeta
	middleware []gin.HandlerFunc
}

// New 创建新的 HTTP 服务实例
func New() *Server {
	// 禁用 gin 的默认日志输出
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	// 设置为生产模式
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		Engine:   gin.New(),
		config:   NewConfig(),
		metadata: make(map[string]*HandlerMeta),
	}
	// 添加默认中间件
	s.Use(internalMiddlewareServerTracing())

	return s
}

// Run 启动 HTTP 服务
func (s *Server) Run() error {
	s.doPrintRoute()

	srv := &http.Server{
		Addr:           s.config.Address,
		Handler:        s.Engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}
	return srv.ListenAndServe()
}

func (s *Server) doPrintRoute() {
	// 打印服务信息
	fmt.Printf("\n")
	s.Logger().Infof(context.Background(), "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)
	// 打印路由信息
	fmt.Printf("%s\n", strings.Repeat("-", 60))
	fmt.Printf("%-10s | %-7s | %-15s \n", "ADDRESS", "METHOD", "ROUTE")

	routes := s.Engine.Routes()
	for _, route := range routes {
		fmt.Printf("%-10s | %-7s | %-15s \n",
			s.config.Address,
			route.Method,
			route.Path,
		)
	}
	fmt.Printf("%s\n\n", strings.Repeat("-", 60))
}
