package mhttp

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mingzaily/maltose/os/mlog"
)

// contextKey 定义上下文键类型
type contextKey string

const (
	// requestKey 用于在上下文中存储 Request 对象的键
	requestKey contextKey = "MaltoseRequest"
)

// Request 请求封装
type Request struct {
	*gin.Context
	Server          *Server // 服务器实例
	handlerResponse interface{}
}

// SetHandlerResponse 设置处理函数响应
func (r *Request) SetHandlerResponse(response interface{}) {
	r.handlerResponse = response
}

// GetHandlerResponse 获取处理函数响应
func (r *Request) GetHandlerResponse() interface{} {
	return r.handlerResponse
}

// GetServerName 获取服务器名称
func (r *Request) GetServerName() string {
	return r.Server.GetServerName()
}

// Logger 获取日志实例
func (r *Request) Logger() *mlog.Logger {
	return r.Server.Logger()
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

// WithRequest 将 Request 对象存储到上下文中
func WithRequest(ctx context.Context, r *Request) context.Context {
	return context.WithValue(ctx, requestKey, r)
}
