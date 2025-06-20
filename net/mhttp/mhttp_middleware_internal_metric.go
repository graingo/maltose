package mhttp

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/os/mmetric"
	"go.opentelemetry.io/otel/attribute"
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
		ctx                       = r.Request.Context()
		serverAddress, serverPort = parseHostPort(r.Request.Host)
		httpRoute                 = r.FullPath()
		protocolVersion           string
		requestAttributes         []attribute.KeyValue
	)

	if httpRoute == "" {
		httpRoute = r.Request.URL.Path
	}
	if array := strings.Split(r.Request.Proto, "/"); len(array) > 1 {
		protocolVersion = array[1]
	}
	if localAddr := r.Request.Context().Value(http.LocalAddrContextKey); localAddr != nil {
		_, serverPort = parseHostPort(localAddr.(net.Addr).String())
	}

	requestAttributes = []attribute.KeyValue{
		attribute.String(metricAttrKeyServerAddress, serverAddress),
		attribute.String(metricAttrKeyServerPort, serverPort),
		attribute.String(metricAttrKeyHttpRoute, httpRoute),
		attribute.String(metricAttrKeyUrlSchema, getSchema(r)),
		attribute.String(metricAttrKeyHttpRequestMethod, r.Request.Method),
		attribute.String(metricAttrKeyNetworkProtocolVersion, protocolVersion),
	}
	options := mmetric.WithAttributes(requestAttributes...)

	// increase active request count
	metricManager.HttpServerRequestActive.Inc(ctx, options)

	// record request body size
	metricManager.HttpServerRequestBodySize.Add(
		ctx,
		float64(r.Request.ContentLength),
		options,
	)
}

// handle metrics after request done
func (s *Server) handleMetricsAfterRequestDone(r *Request, startTime time.Time) {
	if !mmetric.IsEnabled() {
		return
	}

	var (
		ctx                       = r.Request.Context()
		durationMilli             = float64(time.Since(startTime).Milliseconds())
		serverAddress, serverPort = parseHostPort(r.Request.Host)
		httpRoute                 = r.FullPath()
		protocolVersion           string
	)

	if httpRoute == "" {
		httpRoute = r.Request.URL.Path
	}
	if array := strings.Split(r.Request.Proto, "/"); len(array) > 1 {
		protocolVersion = array[1]
	}
	if localAddr := r.Request.Context().Value(http.LocalAddrContextKey); localAddr != nil {
		_, serverPort = parseHostPort(localAddr.(net.Addr).String())
	}

	requestAttributes := []attribute.KeyValue{
		attribute.String(metricAttrKeyServerAddress, serverAddress),
		attribute.String(metricAttrKeyServerPort, serverPort),
		attribute.String(metricAttrKeyHttpRoute, httpRoute),
		attribute.String(metricAttrKeyUrlSchema, getSchema(r)),
		attribute.String(metricAttrKeyHttpRequestMethod, r.Request.Method),
		attribute.String(metricAttrKeyNetworkProtocolVersion, protocolVersion),
	}
	requestOptions := mmetric.WithAttributes(requestAttributes...)

	responseAttributes := make([]attribute.KeyValue, 0, len(requestAttributes)+2)
	responseAttributes = append(responseAttributes, requestAttributes...)
	responseAttributes = append(responseAttributes, attribute.Int(metricAttrKeyHttpResponseStatusCode, r.Writer.Status()))

	if len(r.Errors) > 0 {
		var errCode int
		if err := r.Errors[0]; err != nil {
			errCode = merror.Code(err).Code()
		}
		responseAttributes = append(responseAttributes, attribute.Int(metricAttrKeyErrorCode, errCode))
	}
	responseOptions := mmetric.WithAttributes(responseAttributes...)

	// increase request total
	metricManager.HttpServerRequestTotal.Inc(ctx, responseOptions)

	// decrease active request count
	metricManager.HttpServerRequestActive.Dec(ctx, requestOptions)

	// record response body size
	metricManager.HttpServerResponseBodySize.Add(
		ctx,
		float64(r.Writer.Size()),
		responseOptions,
	)

	// record request duration total
	metricManager.HttpServerRequestDurationTotal.Add(
		ctx,
		durationMilli,
		responseOptions,
	)

	// record request duration distribution
	metricManager.HttpServerRequestDuration.Record(
		durationMilli,
		responseOptions,
	)
}
