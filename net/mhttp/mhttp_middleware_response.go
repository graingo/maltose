package mhttp

import (
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
		r.Next()

		// 如果已经写入了响应,则跳过
		if r.Writer.Written() {
			return
		}

		var (
			msg  string
			code = mcode.Success
			data = r.GetHandlerResponse()
		)

		// 处理错误情况
		if len(r.Errors) > 0 {
			err := r.Errors.Last().Err
			if merr, ok := err.(*merror.Error); ok {
				code = merr.Code()
				msg = merr.Error()
			} else {
				code = mcode.InternalError
				msg = err.Error()
			}
			data = nil
		} else if status := r.Writer.Status(); status != http.StatusOK {
			// 处理 HTTP 状态码错误
			msg = http.StatusText(status)
			switch status {
			case http.StatusNotFound:
				code = mcode.NotFound
			case http.StatusForbidden:
				code = mcode.Forbidden
			case http.StatusUnauthorized:
				code = mcode.Unauthorized
			default:
				code = mcode.InternalError
			}
			data = nil
			// 创建错误对象供其他中间件使用
			r.Error(merror.NewCode(code, msg))
		} else {
			msg = "success"
		}

		// 返回标准响应
		r.JSON(r.Writer.Status(), DefaultResponse{
			Code:    code.Code(),
			Message: msg,
			Data:    data,
		})
	}
}
