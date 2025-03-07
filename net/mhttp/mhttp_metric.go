package mhttp

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mmetric"
)

const (
	metricAttrKeyServerAddress          = "server.address"
	metricAttrKeyServerPort             = "server.port"
	metricAttrKeyHttpRoute              = "http.route"
	metricAttrKeyUrlSchema              = "url.schema"
	metricAttrKeyHttpRequestMethod      = "http.request.method"
	metricAttrKeyErrorCode              = "error.code"
	metricAttrKeyHttpResponseStatusCode = "http.response.status_code"
	metricAttrKeyNetworkProtocolVersion = "network.protocol.version"
)

// 本地指标管理器
type localMetricManager struct {
	HttpServerRequestActive        mmetric.UpDownCounter
	HttpServerRequestTotal         mmetric.Counter
	HttpServerRequestDuration      mmetric.Histogram
	HttpServerRequestDurationTotal mmetric.Counter
	HttpServerRequestBodySize      mmetric.Counter
	HttpServerResponseBodySize     mmetric.Counter
}

// 全局指标管理器
var metricManager = newMetricManager()

// 创建新的指标管理器
func newMetricManager() *localMetricManager {
	meter := mmetric.GetGlobalProvider().Meter(mmetric.MeterOption{
		Instrument:        instrumentName,
		InstrumentVersion: "v1.0.0",
	})
	mm := &localMetricManager{
		HttpServerRequestDuration: meter.MustHistogram(
			"http.server.request.duration",
			mmetric.MetricOption{
				Help: "请求处理时长",
				Unit: "ms",
				Buckets: []float64{
					1, 5, 10, 25, 50, 75, 100, 250, 500, 750,
					1000, 2500, 5000, 7500, 10000, 30000, 60000,
				},
			},
		),
		HttpServerRequestTotal: meter.MustCounter(
			"http.server.request.total",
			mmetric.MetricOption{
				Help: "请求总数",
				Unit: "",
			},
		),
		HttpServerRequestActive: meter.MustUpDownCounter(
			"http.server.request.active",
			mmetric.MetricOption{
				Help: "活跃请求数",
				Unit: "",
			},
		),
		HttpServerRequestDurationTotal: meter.MustCounter(
			"http.server.request.duration_total",
			mmetric.MetricOption{
				Help: "请求处理总时长",
				Unit: "ms",
			},
		),
		HttpServerRequestBodySize: meter.MustCounter(
			"http.server.request.body_size",
			mmetric.MetricOption{
				Help: "请求体总大小",
				Unit: "bytes",
			},
		),
		HttpServerResponseBodySize: meter.MustCounter(
			"http.server.response.body_size",
			mmetric.MetricOption{
				Help: "响应体总大小",
				Unit: "bytes",
			},
		),
	}
	return mm
}

// 获取请求时长指标选项
func (m *localMetricManager) GetMetricOptionForRequestDurationByMap(attrMap mmetric.AttributeMap) mmetric.Option {
	return mmetric.Option{
		Attributes: attrMap.Pick(
			metricAttrKeyServerAddress,
			metricAttrKeyServerPort,
		),
	}
}

// 获取请求指标选项
func (m *localMetricManager) GetMetricOptionForRequestByMap(attrMap mmetric.AttributeMap) mmetric.Option {
	return mmetric.Option{
		Attributes: attrMap.Pick(
			metricAttrKeyServerAddress,
			metricAttrKeyServerPort,
			metricAttrKeyHttpRoute,
			metricAttrKeyUrlSchema,
			metricAttrKeyHttpRequestMethod,
			metricAttrKeyNetworkProtocolVersion,
		),
	}
}

// 获取响应指标选项
func (m *localMetricManager) GetMetricOptionForResponseByMap(attrMap mmetric.AttributeMap) mmetric.Option {
	return mmetric.Option{
		Attributes: attrMap.Pick(
			metricAttrKeyServerAddress,
			metricAttrKeyServerPort,
			metricAttrKeyHttpRoute,
			metricAttrKeyUrlSchema,
			metricAttrKeyHttpRequestMethod,
			metricAttrKeyNetworkProtocolVersion,
			metricAttrKeyErrorCode,
			metricAttrKeyHttpResponseStatusCode,
		),
	}
}

// 获取指标属性映射
func (m *localMetricManager) GetMetricAttributeMap(r *Request) mmetric.AttributeMap {
	var (
		serverAddress   string
		serverPort      string
		httpRoute       string
		protocolVersion string
		attrMap         = make(mmetric.AttributeMap)
	)

	// 解析服务器地址和端口
	serverAddress, serverPort = parseHostPort(r.Request.Host)
	if localAddr := r.Request.Context().Value(http.LocalAddrContextKey); localAddr != nil {
		_, serverPort = parseHostPort(localAddr.(net.Addr).String())
	}

	// 获取路由路径
	httpRoute = r.FullPath()
	if httpRoute == "" {
		httpRoute = r.Request.URL.Path
	}

	// 获取协议版本
	if array := strings.Split(r.Request.Proto, "/"); len(array) > 1 {
		protocolVersion = array[1]
	}

	// 设置基本属性
	attrMap.Sets(mmetric.AttributeMap{
		metricAttrKeyServerAddress:          serverAddress,
		metricAttrKeyServerPort:             serverPort,
		metricAttrKeyHttpRoute:              httpRoute,
		metricAttrKeyUrlSchema:              getSchema(r),
		metricAttrKeyHttpRequestMethod:      r.Request.Method,
		metricAttrKeyNetworkProtocolVersion: protocolVersion,
	})

	// 设置响应相关属性
	if len(r.Errors) > 0 {
		var errCode int
		if err := r.Errors[0]; err != nil {
			errCode = merror.Code(err).Code()
		}
		attrMap.Sets(mmetric.AttributeMap{
			metricAttrKeyErrorCode:              errCode,
			metricAttrKeyHttpResponseStatusCode: r.Writer.Status(),
		})
	}

	return attrMap
}

// 解析主机和端口
func parseHostPort(hostPort string) (host, port string) {
	parts := strings.Split(hostPort, ":")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], "80"
}

// 获取请求协议
func getSchema(r *Request) string {
	if r.Request.TLS != nil {
		return "https"
	}
	return "http"
}

// 处理请求前的指标收集
func (s *Server) handleMetricsBeforeRequest(r *Request) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx           = r.Request.Context()
		attrMap       = metricManager.GetMetricAttributeMap(r)
		requestOption = metricManager.GetMetricOptionForRequestByMap(attrMap)
	)

	// 增加活跃请求计数
	metricManager.HttpServerRequestActive.Inc(
		ctx,
		requestOption,
	)

	// 记录请求体大小
	metricManager.HttpServerRequestBodySize.Add(
		ctx,
		float64(r.Request.ContentLength),
		requestOption,
	)
}

// 处理请求后的指标收集
func (s *Server) handleMetricsAfterRequestDone(r *Request, startTime time.Time) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx             = r.Request.Context()
		attrMap         = metricManager.GetMetricAttributeMap(r)
		durationMilli   = float64(time.Since(startTime).Milliseconds())
		responseOption  = metricManager.GetMetricOptionForResponseByMap(attrMap)
		histogramOption = metricManager.GetMetricOptionForRequestDurationByMap(attrMap)
	)

	// 增加请求总数
	metricManager.HttpServerRequestTotal.Inc(ctx, responseOption)

	// 减少活跃请求计数
	metricManager.HttpServerRequestActive.Dec(
		ctx,
		metricManager.GetMetricOptionForRequestByMap(attrMap),
	)

	// 记录响应体大小
	metricManager.HttpServerResponseBodySize.Add(
		ctx,
		float64(r.Writer.Size()),
		responseOption,
	)

	// 记录请求处理时长
	metricManager.HttpServerRequestDurationTotal.Add(
		ctx,
		durationMilli,
		responseOption,
	)

	// 记录请求处理时长分布
	metricManager.HttpServerRequestDuration.Record(
		durationMilli,
		histogramOption,
	)
}
