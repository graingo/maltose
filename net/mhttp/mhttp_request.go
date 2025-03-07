package mhttp

import (
	"context"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/graingo/maltose/os/mlog"
)

// contextKey 定义上下文键类型
type contextKey string

const (
	// requestKey 用于在上下文中存储 Request 对象的键
	requestKey  contextKey = "MaltoseRequest"
	ResponseKey contextKey = "MaltoseResponse"
)

// Request 请求封装
type Request struct {
	*gin.Context
	server *Server // 服务器实例
}

// RequestFromCtx 从上下文中获取 Request 对象
func RequestFromCtx(ctx context.Context) *Request {
	if ctx == nil {
		return nil
	}
	if v := ctx.Value(requestKey); v != nil {
		if r, ok := v.(*Request); ok {
			return r
		}
	}
	return nil
}

func newRequest(c *gin.Context, s *Server) *Request {
	// 先尝试从上下文获取
	if r := RequestFromCtx(c.Request.Context()); r != nil {
		return r
	}
	// 创建新的 Request 对象
	r := &Request{Context: c, server: s}
	// 直接修改原始 context，而不是创建新的 request
	r.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestKey, r))
	return r
}

// GetServerName 获取服务器名称
func (r *Request) GetServerName() string {
	return r.server.config.ServerName
}

// Logger 获取日志实例
func (r *Request) Logger() *mlog.Logger {
	return r.server.Logger()
}

// GetHandlerResponse 获取处理器响应
func (r *Request) GetHandlerResponse() any {
	res, _ := r.Get(string(ResponseKey))
	return res
}

// SetHandlerResponse 设置处理器响应
func (r *Request) SetHandlerResponse(res any) {
	r.Set(string(ResponseKey), res)
}

// Error 添加错误信息
func (r *Request) Error(err error) *Request {
	r.Errors = append(r.Errors, &gin.Error{
		Err:  err,
		Type: gin.ErrorTypePrivate,
	})
	return r
}

// GetTranslator 获取翻译器
func (r *Request) GetTranslator() ut.Translator {
	return r.server.translator
}
