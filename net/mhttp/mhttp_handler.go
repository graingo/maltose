package mhttp

import (
	"github.com/gin-gonic/gin"
)

// HandlerFunc 定义基础处理函数类型
type HandlerFunc func(*Request)

// BindHandler 绑定简单处理函数
func (s *Server) BindHandler(method, path string, handler HandlerFunc) {
	s.Handle(method, path, func(c *gin.Context) {
		handler(RequestFromCtx(c))
	})
}

// Bind 绑定对象到根路由
func (s *Server) Bind(object any) {
	s.bindObject(nil, object)
}
