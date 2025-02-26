package mhttp

import (
	"fmt"
	"net/http"

	"github.com/mingzaily/maltose/errors/mcode"
	"github.com/mingzaily/maltose/errors/merror"
)

// DefaultResponse 标准响应结构
type DefaultResponse struct {
	Code    int    `json:"code"`    // 业务码
	Message string `json:"message"` // 提示信息
	Data    any    `json:"data"`    // 业务数据
}

// MiddlewareResponse 标准响应中间件
func MiddlewareResponse() MiddlewareFunc {
	return func(r *Request) {
		// 执行后续中间件
		r.Next()

		// 如果已经写入了响应,则跳过
		if r.Writer.Written() {
			fmt.Println("response has been written")
			return
		}
		fmt.Println("response not been written")

		var response DefaultResponse

		if r.Writer.Status() != http.StatusOK {
			response = DefaultResponse{
				Code:    r.Writer.Status(),
				Message: http.StatusText(r.Writer.Status()),
				Data:    nil,
			}
		} else if len(r.Errors) > 0 {
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

		r.JSON(r.Writer.Status(), response)
	}
}
