package mhttp

import (
	"time"
)

// MiddlewareLog 日志中间件
func MiddlewareLog() MiddlewareFunc {
	return func(r *Request) {
		// 开始时间
		start := time.Now()

		// 获取请求信息
		path := r.Request.URL.Path
		raw := r.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// 执行后续中间件
		r.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 获取响应状态
		status := r.Writer.Status()

		// 记录日志
		r.Logger().Infof(r.Request.Context(),
			"[HTTP] %3d | %7v | %15s | %-7s | %s",
			status,
			latency,
			r.ClientIP(),
			r.Request.Method,
			path,
		)

		// 如果有错误，记录错误日志
		if len(r.Errors) > 0 {
			for _, e := range r.Errors {
				r.Logger().Errorf(r.Request.Context(),
					"[HTTP] %s | Error: %v",
					r.GetServerName(),
					e.Err,
				)
			}
		}
	}
}
