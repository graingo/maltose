package mhttp

import "github.com/mingzaily/maltose/errors/merror"

// MiddlewareError 错误处理中间件
func MiddlewareError() MiddlewareFunc {
	return func(r *Request) {
		defer func() {
			if err := recover(); err != nil {
				// 处理 panic
				r.Logger().Errorf(r.Request.Context(), "Panic recovered: %v", err)
				r.AbortWithError(500, merror.New("Internal Server Error"))
			}
		}()
		r.Next()
	}
}
