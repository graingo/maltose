package mhttp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/graingo/maltose/net/mtrace"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentName                        = "github.com/graingo/maltose/net/mhttp.Server"
	tracingEventHttpRequest               = "http.request"
	tracingEventHttpRequestUrl            = "http.request.url"
	tracingEventHttpHeaders               = "http.headers"
	tracingEventHttpBaggage               = "http.baggage"
	tracingEventHttpResponse              = "http.response"
	tracingMiddlewareHandled   contextKey = "TracingMiddlewareHandled"
)

// internalMiddlewareDefaultResponse 内部默认响应处理中间件
func internalMiddlewareDefaultResponse() MiddlewareFunc {
	return func(r *Request) {
		r.Next()

		// 如果已经写入了响应,则跳过
		if r.Writer.Written() {
			return
		}

		// 处理错误情况
		if len(r.Errors) > 0 {
			err := r.Errors.Last().Err
			r.String(500, fmt.Sprintf("Error: %s", err.Error()))
			return
		}

		// 获取处理器响应
		if res := r.GetHandlerResponse(); res != nil {
			switch v := res.(type) {
			case string:
				r.String(200, v)
			case []byte:
				r.String(200, string(v))
			default:
				r.String(200, fmt.Sprintf("%v", v))
			}
			return
		}

		// 没有响应则返回空字符串
		r.String(200, "")
	}
}

// internalMiddlewareRecovery 内部错误恢复中间件
func internalMiddlewareRecovery() MiddlewareFunc {
	return func(r *Request) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				r.Logger().Errorf(r.Request.Context(), "Panic recovered: %v", err)
				// 返回 500 错误
				r.String(500, "Internal Server Error")
			}
		}()
		r.Next()
	}
}

// internalMiddlewareMetric 内部指标收集中间件
func internalMiddlewareMetric() MiddlewareFunc {
	return func(r *Request) {
		// 记录开始时间
		startTime := time.Now()

		// 请求前指标收集
		r.server.handleMetricsBeforeRequest(r)

		// 执行后续中间件
		r.Next()

		// 请求后指标收集
		r.server.handleMetricsAfterRequestDone(r, startTime)
	}
}

// internalMiddlewareServerTrace 返回一个中间件用于OpenTelemetry跟踪
func internalMiddlewareServerTrace() MiddlewareFunc {
	return func(r *Request) {
		ctx := r.Request.Context()

		// 避免重复处理
		if ctx.Value(tracingMiddlewareHandled) != nil {
			r.Next()
			return
		}

		ctx = context.WithValue(ctx, tracingMiddlewareHandled, 1)

		// 如果使用默认provider则跳过复杂的tracing
		if mtrace.IsUsingDefaultProvider() {
			r.Next()
			return
		}

		// 创建tracer
		tr := otel.GetTracerProvider().Tracer(
			instrumentName,
			trace.WithInstrumentationVersion("v1.0.0"),
		)

		// 提取上下文和baggage
		ctx = otel.GetTextMapPropagator().Extract(
			ctx,
			propagation.HeaderCarrier(r.Request.Header),
		)

		// 规范化操作名称: HTTP {method} {route}
		spanName := fmt.Sprintf("HTTP %s %s", r.Request.Method, r.FullPath())
		if r.FullPath() == "" {
			spanName = fmt.Sprintf("HTTP %s %s", r.Request.Method, r.Request.URL.Path)
		}

		// 创建span
		ctx, span := tr.Start(
			ctx,
			spanName,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// 添加请求事件
		span.AddEvent(tracingEventHttpRequest, trace.WithAttributes(
			attribute.String(tracingEventHttpRequestUrl, r.Request.URL.String()),
			attribute.String(tracingEventHttpHeaders, cast.ToString(cast.ToStringMap(r.Request.Header))),
			attribute.String(tracingEventHttpBaggage, cast.ToString(mtrace.GetBaggageMap(ctx))),
		))

		// 注入追踪上下文
		r.Request = r.Request.WithContext(ctx)

		// 继续处理请求
		r.Next()

		// 错误处理
		if len(r.Errors) > 0 {
			// 收集所有错误信息
			var errMsgs []string
			for _, err := range r.Errors {
				errMsgs = append(errMsgs, err.Error())
			}
			span.SetStatus(codes.Error, strings.Join(errMsgs, "; "))
		} else {
			span.SetStatus(httpStatusCodeToSpanStatus(r.Writer.Status()))
		}

		// 添加响应事件
		span.AddEvent(tracingEventHttpResponse, trace.WithAttributes(
			attribute.Int("http.status_code", r.Writer.Status()),
			attribute.Int("http.response_content_length", r.Writer.Size()),
			attribute.String(tracingEventHttpHeaders, fmt.Sprint(r.Writer.Header())),
		))
	}
}

// httpStatusCodeToSpanStatus 将 HTTP 状态码转换为 span 状态
func httpStatusCodeToSpanStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 400 {
		return codes.Error, fmt.Sprintf("HTTP %d", code)
	}
	return codes.Ok, ""
}
