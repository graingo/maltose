package mhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DefaultResponse 标准响应结构
type DefaultResponse struct {
	Code    int         `json:"code"`    // 业务码
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 业务数据
}

// MiddlewareResponse 标准响应中间件
func MiddlewareResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		r := &Request{Context: c}

		// 执行后续中间件
		c.Next()

		// 如果已经写入了响应,则跳过
		if c.Writer.Written() {
			return
		}

		// 获取错误信息
		var response DefaultResponse
		if len(c.Errors) > 0 {
			// 处理错误情况
			err := c.Errors.Last()
			response = DefaultResponse{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
				Data:    nil,
			}
		} else {
			// 正常响应
			response = DefaultResponse{
				Code:    http.StatusOK,
				Message: "success",
				Data:    r.GetHandlerResponse(),
			}
		}

		c.JSON(c.Writer.Status(), response)
	}
}
