package mhttp

import "time"

// internalMiddlewareMetric 内部指标收集中间件
func internalMiddlewareMetric() MiddlewareFunc {
	return func(r *Request) {
		// 记录开始时间
		startTime := time.Now()

		// 请求前指标收集
		r.Server.handleMetricsBeforeRequest(r)

		// 执行后续中间件
		r.Next()

		// 请求后指标收集
		r.Server.handleMetricsAfterRequestDone(r, startTime)
	}
}
