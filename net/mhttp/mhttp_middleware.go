package mhttp

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// MiddlewareFunc 定义中间件函数类型
type MiddlewareFunc func(*Request)

// Use 添加全局中间件
func (s *Server) Use(middleware ...MiddlewareFunc) {
	// 转换为 gin 的中间件格式
	for _, m := range middleware {
		handler := func(c *gin.Context) {
			r := newRequest(c, s)
			m(r)
		}
		s.middlewares = append(s.middlewares, m)
		s.Engine.Use(handler)
	}
}

// internalMiddlewareDefaultResponse 内部默认响应处理中间件
func internalMiddlewareDefaultResponse() MiddlewareFunc {
	return func(r *Request) {
		r.Next()

		// 如果已经写入了响应,则跳过
		if r.Writer.Written() {
			return
		}

		// 处理错误情况
		if len(r.Errors) > 0 {
			err := r.Errors.Last().Err
			r.String(500, fmt.Sprintf("Error: %s", err.Error()))
			return
		}

		// 获取处理器响应
		if res := r.GetHandlerResponse(); res != nil {
			switch v := res.(type) {
			case string:
				r.String(200, v)
			case []byte:
				r.String(200, string(v))
			default:
				r.String(200, fmt.Sprintf("%v", v))
			}
			return
		}

		// 没有响应则返回空字符串
		r.String(200, "")
	}
}

func internalMiddlewareRecovery() MiddlewareFunc {
	return func(r *Request) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				r.Logger().Errorf(r.Request.Context(), "Panic recovered: %v", err)
				// 返回 500 错误
				r.String(500, "Internal Server Error")
			}
		}()
		r.Next()
	}
}
