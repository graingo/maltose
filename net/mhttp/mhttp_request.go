package mhttp

import (
	"github.com/gin-gonic/gin"
)

// Request 请求封装
type Request struct {
	*gin.Context
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
