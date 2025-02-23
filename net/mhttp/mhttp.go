package mhttp

import (
	"context"
	"fmt"
	"io"
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
	config   ServerConfig
	metadata map[string]*HandlerMeta
}

// New 创建新的 HTTP 服务实例
func New() *Server {
	// 禁用 gin 的默认日志输出
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	// 设置为生产模式
	gin.SetMode(gin.ReleaseMode)

	g := gin.New()

	// 添加默认中间件
	g.Use(internalMiddlewareServerTracing())

	return &Server{
		Engine:   g,
		config:   NewConfig(),
		metadata: make(map[string]*HandlerMeta),
	}
}

// Run 启动 HTTP 服务
func (s *Server) Run() error {
	s.Logger().Infof(context.Background(), "HTTP server %s is running on %s", s.config.ServerName, s.config.Address)

	// 打印路由信息
	fmt.Printf("\n%s\n", strings.Repeat("-", 120))
	fmt.Printf("%-10s | %-7s | %-15s \n", "ADDRESS", "METHOD", "ROUTE")

	routes := s.Engine.Routes()
	for _, route := range routes {
		fmt.Printf("%-10s | %-7s | %-15s \n",
			s.config.Address,
			route.Method,
			route.Path,
		)
	}
	fmt.Printf("%s\n\n", strings.Repeat("-", 120))
	return s.Engine.Run(s.config.Address)
}
