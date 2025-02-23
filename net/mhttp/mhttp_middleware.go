package mhttp

import (
	"github.com/gin-gonic/gin"
)

// MiddlewareManager 中间件管理器
type MiddlewareManager struct {
	handlers []gin.HandlerFunc
}

// Use 添加全局中间件
func (s *Server) Use(middleware ...gin.HandlerFunc) {
	s.Engine.Use(middleware...)
}
