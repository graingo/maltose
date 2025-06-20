package mclient

import (
	"net/http"
	"time"

	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/attribute"
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

// handle metrics before request
func handleMetricsBeforeRequest(req *http.Request) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx        = req.Context()
		attributes = []attribute.KeyValue{
			attribute.String(metricAttrKeyUrlHost, getHost(req.URL)),
			attribute.String(metricAttrKeyUrlPath, getPath(req.URL)),
			attribute.String(metricAttrKeyUrlSchema, getSchema(req.URL)),
			attribute.String(metricAttrKeyHttpRequestMethod, req.Method),
			attribute.String(metricAttrKeyNetworkProtocolVersion, getProtocolVersion(req.Proto)),
		}
	)

	metricManager.HttpClientRequestBodySize.Add(
		ctx,
		float64(req.ContentLength),
		mmetric.WithAttributes(attributes...),
	)
}

// handle metrics after request done
func handleMetricsAfterRequestDone(req *http.Request, resp *http.Response, err error, startTime time.Time) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx           = req.Context()
		durationMilli = float64(time.Since(startTime).Milliseconds())
		attributes    = make([]attribute.KeyValue, 0, 7)
	)

	attributes = append(attributes,
		attribute.String(metricAttrKeyUrlHost, getHost(req.URL)),
		attribute.String(metricAttrKeyUrlPath, getPath(req.URL)),
		attribute.String(metricAttrKeyUrlSchema, getSchema(req.URL)),
		attribute.String(metricAttrKeyHttpRequestMethod, req.Method),
		attribute.String(metricAttrKeyNetworkProtocolVersion, getProtocolVersion(req.Proto)),
	)

	if resp != nil {
		attributes = append(attributes, attribute.Int(metricAttrKeyHttpResponseStatusCode, resp.StatusCode))
	}

	if err != nil {
		attributes = append(attributes, attribute.Int(metricAttrKeyErrorCode, 1))
	}

	metricManager.HttpClientRequestTotal.Inc(ctx, mmetric.WithAttributes(attributes...))
	metricManager.HttpClientRequestDuration.Record(durationMilli, mmetric.WithAttributes(attributes...))
	metricManager.HttpClientRequestDurationTotal.Add(ctx, durationMilli, mmetric.WithAttributes(attributes...))

	if resp != nil {
		metricManager.HttpClientResponseBodySize.Add(
			ctx,
			float64(resp.ContentLength),
			mmetric.WithAttributes(attributes...),
		)
	}

	if err != nil {
		metricManager.HttpClientErrorTotal.Inc(ctx, mmetric.WithAttributes(attributes...))
	}
}
