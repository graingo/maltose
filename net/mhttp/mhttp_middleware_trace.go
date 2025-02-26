package mhttp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mingzaily/maltose/net/mtrace"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentName                        = "github.com/mingzaily/maltose/net/mhttp.Server"
	tracingEventHttpRequest               = "http.request"
	tracingEventHttpRequestUrl            = "http.request.url"
	tracingEventHttpHeaders               = "http.headers"
	tracingEventHttpBaggage               = "http.baggage"
	tracingEventHttpResponse              = "http.response"
	tracingMiddlewareHandled   contextKey = "TracingMiddlewareHandled"
)

// internalMiddlewareServerTracing 返回一个中间件用于OpenTelemetry跟踪
func internalMiddlewareServerTracing() MiddlewareFunc {
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
