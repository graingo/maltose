package mclient

import (
	"net/http"
	"time"

	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/attribute"
)

// localMetricManager is the local metric manager.
type localMetricManager struct {
	HTTPClientRequestTotal         mmetric.Counter
	HTTPClientRequestDuration      mmetric.Histogram
	HTTPClientRequestDurationTotal mmetric.Counter
	HTTPClientRequestBodySize      mmetric.Counter
	HTTPClientResponseBodySize     mmetric.Counter
	HTTPpClientErrorTotal          mmetric.Counter
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
		HTTPClientRequestDuration: meter.MustHistogram(
			"http.client.request.duration",
			mmetric.MetricOption{
				Help: "request duration",
				Unit: "ms",
			},
		),
		HTTPClientRequestTotal: meter.MustCounter(
			"http.client.request.total",
			mmetric.MetricOption{
				Help: "request total",
				Unit: "",
			},
		),
		HTTPClientRequestDurationTotal: meter.MustCounter(
			"http.client.request.duration_total",
			mmetric.MetricOption{
				Help: "request duration total",
				Unit: "ms",
			},
		),
		HTTPClientRequestBodySize: meter.MustCounter(
			"http.client.request.body_size",
			mmetric.MetricOption{
				Help: "request body size",
				Unit: "bytes",
			},
		),
		HTTPClientResponseBodySize: meter.MustCounter(
			"http.client.response.body_size",
			mmetric.MetricOption{
				Help: "response body size",
				Unit: "bytes",
			},
		),
		HTTPpClientErrorTotal: meter.MustCounter(
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
	var (
		ctx        = req.Context()
		attributes = []attribute.KeyValue{
			attribute.String(mmetric.AttrURLHost, getHost(req.URL)),
			attribute.String(mmetric.AttrURLPath, getPath(req.URL)),
			attribute.String(mmetric.AttrURLScheme, getSchema(req.URL)),
			attribute.String(mmetric.AttrHTTPRequestMethod, req.Method),
			attribute.String(mmetric.AttrNetworkProtocolVersion, getProtocolVersion(req.Proto)),
		}
	)

	metricManager.HTTPClientRequestBodySize.Add(
		ctx,
		float64(req.ContentLength),
		mmetric.WithAttributes(attributes...),
	)
}

// handle metrics after request done
func handleMetricsAfterRequestDone(req *http.Request, resp *http.Response, err error, startTime time.Time) {
	var (
		ctx           = req.Context()
		durationMilli = float64(time.Since(startTime).Milliseconds())
		attributes    = make([]attribute.KeyValue, 0, 7)
	)

	attributes = append(attributes,
		attribute.String(mmetric.AttrURLHost, getHost(req.URL)),
		attribute.String(mmetric.AttrURLPath, getPath(req.URL)),
		attribute.String(mmetric.AttrURLScheme, getSchema(req.URL)),
		attribute.String(mmetric.AttrHTTPRequestMethod, req.Method),
		attribute.String(mmetric.AttrNetworkProtocolVersion, getProtocolVersion(req.Proto)),
	)

	if resp != nil {
		attributes = append(attributes, attribute.Int(mmetric.AttrHTTPResponseStatusCode, resp.StatusCode))
	}

	if err != nil {
		attributes = append(attributes, attribute.Int(mmetric.AttrErrorCode, 1))
	}

	metricManager.HTTPClientRequestTotal.Inc(ctx, mmetric.WithAttributes(attributes...))
	metricManager.HTTPClientRequestDuration.Record(durationMilli, mmetric.WithAttributes(attributes...))
	metricManager.HTTPClientRequestDurationTotal.Add(ctx, durationMilli, mmetric.WithAttributes(attributes...))

	if resp != nil {
		metricManager.HTTPClientResponseBodySize.Add(
			ctx,
			float64(resp.ContentLength),
			mmetric.WithAttributes(attributes...),
		)
	}

	if err != nil {
		metricManager.HTTPpClientErrorTotal.Inc(ctx, mmetric.WithAttributes(attributes...))
	}
}
