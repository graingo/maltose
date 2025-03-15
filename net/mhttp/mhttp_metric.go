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

// localMetricManager is the local metric manager.
type localMetricManager struct {
	HttpServerRequestActive        mmetric.UpDownCounter
	HttpServerRequestTotal         mmetric.Counter
	HttpServerRequestDuration      mmetric.Histogram
	HttpServerRequestDurationTotal mmetric.Counter
	HttpServerRequestBodySize      mmetric.Counter
	HttpServerResponseBodySize     mmetric.Counter
}

// global metric manager
var metricManager = newMetricManager()

// create new metric manager
func newMetricManager() *localMetricManager {
	meter := mmetric.GetProvider().Meter(mmetric.MeterOption{
		Instrument:        instrumentName,
		InstrumentVersion: "v1.0.0",
	})
	mm := &localMetricManager{
		HttpServerRequestDuration: meter.MustHistogram(
			"http.server.request.duration",
			mmetric.MetricOption{
				Help: "request duration",
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
				Help: "request total",
				Unit: "",
			},
		),
		HttpServerRequestActive: meter.MustUpDownCounter(
			"http.server.request.active",
			mmetric.MetricOption{
				Help: "active request",
				Unit: "",
			},
		),
		HttpServerRequestDurationTotal: meter.MustCounter(
			"http.server.request.duration_total",
			mmetric.MetricOption{
				Help: "request duration total",
				Unit: "ms",
			},
		),
		HttpServerRequestBodySize: meter.MustCounter(
			"http.server.request.body_size",
			mmetric.MetricOption{
				Help: "request body size",
				Unit: "bytes",
			},
		),
		HttpServerResponseBodySize: meter.MustCounter(
			"http.server.response.body_size",
			mmetric.MetricOption{
				Help: "response body size",
				Unit: "bytes",
			},
		),
	}
	return mm
}

// get request duration metric option
func (m *localMetricManager) GetMetricOptionForRequestDurationByMap(attrMap mmetric.AttributeMap) mmetric.Option {
	return mmetric.Option{
		Attributes: attrMap.Pick(
			metricAttrKeyServerAddress,
			metricAttrKeyServerPort,
		),
	}
}

// get request metric option
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

// get response metric option
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

// get metric attribute map
func (m *localMetricManager) GetMetricAttributeMap(r *Request) mmetric.AttributeMap {
	var (
		serverAddress   string
		serverPort      string
		httpRoute       string
		protocolVersion string
		attrMap         = make(mmetric.AttributeMap)
	)

	// parse server address and port
	serverAddress, serverPort = parseHostPort(r.Request.Host)
	if localAddr := r.Request.Context().Value(http.LocalAddrContextKey); localAddr != nil {
		_, serverPort = parseHostPort(localAddr.(net.Addr).String())
	}

	// get route path
	httpRoute = r.FullPath()
	if httpRoute == "" {
		httpRoute = r.Request.URL.Path
	}

	// get protocol version
	if array := strings.Split(r.Request.Proto, "/"); len(array) > 1 {
		protocolVersion = array[1]
	}

	// set basic attributes
	attrMap.Sets(mmetric.AttributeMap{
		metricAttrKeyServerAddress:          serverAddress,
		metricAttrKeyServerPort:             serverPort,
		metricAttrKeyHttpRoute:              httpRoute,
		metricAttrKeyUrlSchema:              getSchema(r),
		metricAttrKeyHttpRequestMethod:      r.Request.Method,
		metricAttrKeyNetworkProtocolVersion: protocolVersion,
	})

	// set response related attributes
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

// parse host and port
func parseHostPort(hostPort string) (host, port string) {
	parts := strings.Split(hostPort, ":")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], "80"
}

// get request schema
func getSchema(r *Request) string {
	if r.Request.TLS != nil {
		return "https"
	}
	return "http"
}

// handle metrics before request
func (s *Server) handleMetricsBeforeRequest(r *Request) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx           = r.Request.Context()
		attrMap       = metricManager.GetMetricAttributeMap(r)
		requestOption = metricManager.GetMetricOptionForRequestByMap(attrMap)
	)

	// increase active request count
	metricManager.HttpServerRequestActive.Inc(
		ctx,
		requestOption,
	)

	// record request body size
	metricManager.HttpServerRequestBodySize.Add(
		ctx,
		float64(r.Request.ContentLength),
		requestOption,
	)
}

// handle metrics after request done
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

	// increase request total
	metricManager.HttpServerRequestTotal.Inc(ctx, responseOption)

	// decrease active request count
	metricManager.HttpServerRequestActive.Dec(
		ctx,
		metricManager.GetMetricOptionForRequestByMap(attrMap),
	)

	// record response body size
	metricManager.HttpServerResponseBodySize.Add(
		ctx,
		float64(r.Writer.Size()),
		responseOption,
	)

	// record request duration total
	metricManager.HttpServerRequestDurationTotal.Add(
		ctx,
		durationMilli,
		responseOption,
	)

	// record request duration distribution
	metricManager.HttpServerRequestDuration.Record(
		durationMilli,
		histogramOption,
	)
}
