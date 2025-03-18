package mclient

import (
	"net/http"
	"time"

	"github.com/graingo/maltose/os/mmetric"
)

const (
	metricAttrKeyClientAddress          = "client.address"
	metricAttrKeyUrlHost                = "url.host"
	metricAttrKeyUrlPath                = "url.path"
	metricAttrKeyUrlSchema              = "url.schema"
	metricAttrKeyHttpRequestMethod      = "http.request.method"
	metricAttrKeyErrorCode              = "error.code"
	metricAttrKeyHttpResponseStatusCode = "http.response.status_code"
	metricAttrKeyNetworkProtocolVersion = "network.protocol.version"
)

// localMetricManager is the local metric manager.
type localMetricManager struct {
	HttpClientRequestTotal         mmetric.Counter
	HttpClientRequestDuration      mmetric.Histogram
	HttpClientRequestDurationTotal mmetric.Counter
	HttpClientRequestBodySize      mmetric.Counter
	HttpClientResponseBodySize     mmetric.Counter
	HttpClientErrorTotal           mmetric.Counter
}

// global metric manager
var metricManager = newMetricManager()

// create new metric manager
func newMetricManager() *localMetricManager {
	meter := mmetric.GetProvider().Meter(mmetric.MeterOption{
		Instrument:        instrumentName,
		InstrumentVersion: "v1.0.0",
	})
	return &localMetricManager{
		HttpClientRequestDuration: meter.MustHistogram(
			"http.client.request.duration",
			mmetric.MetricOption{
				Help: "request duration",
				Unit: "ms",
				Buckets: []float64{
					1, 5, 10, 25, 50, 75, 100, 250, 500, 750,
					1000, 2500, 5000, 7500, 10000, 30000, 60000,
				},
			},
		),
		HttpClientRequestTotal: meter.MustCounter(
			"http.client.request.total",
			mmetric.MetricOption{
				Help: "request total",
				Unit: "",
			},
		),
		HttpClientRequestDurationTotal: meter.MustCounter(
			"http.client.request.duration_total",
			mmetric.MetricOption{
				Help: "request duration total",
				Unit: "ms",
			},
		),
		HttpClientRequestBodySize: meter.MustCounter(
			"http.client.request.body_size",
			mmetric.MetricOption{
				Help: "request body size",
				Unit: "bytes",
			},
		),
		HttpClientResponseBodySize: meter.MustCounter(
			"http.client.response.body_size",
			mmetric.MetricOption{
				Help: "response body size",
				Unit: "bytes",
			},
		),
		HttpClientErrorTotal: meter.MustCounter(
			"http.client.error.total",
			mmetric.MetricOption{
				Help: "error total",
				Unit: "",
			},
		),
	}
}

// get request metric option
func (m *localMetricManager) GetMetricOptionForRequestByMap(attrMap mmetric.AttributeMap) mmetric.Option {
	return mmetric.Option{
		Attributes: attrMap.Pick(
			metricAttrKeyUrlHost,
			metricAttrKeyUrlPath,
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
			metricAttrKeyUrlHost,
			metricAttrKeyUrlPath,
			metricAttrKeyUrlSchema,
			metricAttrKeyHttpRequestMethod,
			metricAttrKeyNetworkProtocolVersion,
			metricAttrKeyErrorCode,
			metricAttrKeyHttpResponseStatusCode,
		),
	}
}

// get metric attribute map
func (m *localMetricManager) GetMetricAttributeMap(req *http.Request, resp *http.Response, err error) mmetric.AttributeMap {
	var (
		attrMap = make(mmetric.AttributeMap)
	)

	// Set basic attributes
	attrMap.Sets(mmetric.AttributeMap{
		metricAttrKeyUrlHost:                getHost(req.URL),
		metricAttrKeyUrlPath:                getPath(req.URL),
		metricAttrKeyUrlSchema:              getSchema(req.URL),
		metricAttrKeyHttpRequestMethod:      req.Method,
		metricAttrKeyNetworkProtocolVersion: getProtocolVersion(req.Proto),
	})

	// Set response related attributes
	if resp != nil {
		attrMap.Sets(mmetric.AttributeMap{
			metricAttrKeyHttpResponseStatusCode: resp.StatusCode,
		})
	}

	// Set error related attributes
	if err != nil {
		attrMap.Sets(mmetric.AttributeMap{
			metricAttrKeyErrorCode: 1, // Use a generic error code
		})
	}

	return attrMap
}

// handle metrics before request
func handleMetricsBeforeRequest(req *http.Request) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx           = req.Context()
		attrMap       = metricManager.GetMetricAttributeMap(req, nil, nil)
		requestOption = metricManager.GetMetricOptionForRequestByMap(attrMap)
	)

	// Record request body size
	metricManager.HttpClientRequestBodySize.Add(
		ctx,
		float64(req.ContentLength),
		requestOption,
	)
}

// handle metrics after request done
func handleMetricsAfterRequestDone(req *http.Request, resp *http.Response, err error, startTime time.Time) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx            = req.Context()
		attrMap        = metricManager.GetMetricAttributeMap(req, resp, err)
		durationMilli  = float64(time.Since(startTime).Milliseconds())
		responseOption = metricManager.GetMetricOptionForResponseByMap(attrMap)
	)

	// Increase request total
	metricManager.HttpClientRequestTotal.Inc(ctx, responseOption)

	// Record request duration
	metricManager.HttpClientRequestDuration.Record(ctx, durationMilli, responseOption)
	metricManager.HttpClientRequestDurationTotal.Add(ctx, durationMilli, responseOption)

	// Record response body size if available
	if resp != nil {
		metricManager.HttpClientResponseBodySize.Add(
			ctx,
			float64(resp.ContentLength),
			responseOption,
		)
	}

	// Record error if any
	if err != nil {
		metricManager.HttpClientErrorTotal.Inc(ctx, responseOption)
	}
}
