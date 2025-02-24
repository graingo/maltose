package mhttp

import (
	"fmt"
	"net/http"

	"github.com/mingzaily/maltose/errors/mcode"
	"github.com/mingzaily/maltose/errors/merror"
)

// DefaultResponse 标准响应结构
type DefaultResponse struct {
	Code    int         `json:"code"`    // 业务码
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 业务数据
}

// MiddlewareResponse 标准响应中间件
func MiddlewareResponse() MiddlewareFunc {
	return func(r *Request) {
		// 执行后续中间件
		r.Next()

		// 如果已经写入了响应,则跳过
		if r.Writer.Written() {
			return
		}

		fmt.Printf("r.address: %p, r:%+v\n", r, r)
		// 获取错误信息
		var response DefaultResponse
		if len(r.Errors) > 0 {
			err := r.Errors.Last().Err
			if merr, ok := err.(*merror.Error); ok {
				response = DefaultResponse{
					Code:    merr.Code().Code(),
					Message: merr.Error(),
					Data:    nil,
				}
			} else {
				response = DefaultResponse{
					Code:    mcode.InternalError.Code(),
					Message: err.Error(),
					Data:    nil,
				}
			}
		} else {
			response = DefaultResponse{
				Code:    mcode.Success.Code(),
				Message: "success",
				Data:    r.GetHandlerResponse(),
			}
		}

		r.JSON(http.StatusOK, response)
	}
}
