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

// localMetricManager is the local metric manager.
type localMetricManager struct {
	HTTPServerRequestActive        mmetric.UpDownCounter
	HTTPServerRequestTotal         mmetric.Counter
	HTTPServerRequestDuration      mmetric.Histogram
	HTTPServerRequestDurationTotal mmetric.Counter
	HTTPServerRequestBodySize      mmetric.Counter
	HTTPServerResponseBodySize     mmetric.Counter
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
		HTTPServerRequestDuration: meter.MustHistogram(
			"http.server.request.duration",
			mmetric.MetricOption{
				Help: "request duration",
				Unit: "ms",
			},
		),
		HTTPServerRequestTotal: meter.MustCounter(
			"http.server.request.total",
			mmetric.MetricOption{
				Help: "request total",
				Unit: "",
			},
		),
		HTTPServerRequestActive: meter.MustUpDownCounter(
			"http.server.request.active",
			mmetric.MetricOption{
				Help: "active request",
				Unit: "",
			},
		),
		HTTPServerRequestDurationTotal: meter.MustCounter(
			"http.server.request.duration_total",
			mmetric.MetricOption{
				Help: "request duration total",
				Unit: "ms",
			},
		),
		HTTPServerRequestBodySize: meter.MustCounter(
			"http.server.request.body_size",
			mmetric.MetricOption{
				Help: "request body size",
				Unit: "bytes",
			},
		),
		HTTPServerResponseBodySize: meter.MustCounter(
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
		attribute.String(mmetric.AttrServerAddress, serverAddress),
		attribute.String(mmetric.AttrServerPort, serverPort),
		attribute.String(mmetric.AttrHTTPRoute, httpRoute),
		attribute.String(mmetric.AttrURLScheme, getSchema(r)),
		attribute.String(mmetric.AttrHTTPRequestMethod, r.Request.Method),
		attribute.String(mmetric.AttrNetworkProtocolVersion, protocolVersion),
	}
	options := mmetric.WithAttributes(requestAttributes...)

	// increase active request count
	metricManager.HTTPServerRequestActive.Inc(ctx, options)

	// record request body size
	metricManager.HTTPServerRequestBodySize.Add(
		ctx,
		float64(r.Request.ContentLength),
		options,
	)
}

// handle metrics after request done
func (s *Server) handleMetricsAfterRequestDone(r *Request, startTime time.Time) {
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
		attribute.String(mmetric.AttrServerAddress, serverAddress),
		attribute.String(mmetric.AttrServerPort, serverPort),
		attribute.String(mmetric.AttrHTTPRoute, httpRoute),
		attribute.String(mmetric.AttrURLScheme, getSchema(r)),
		attribute.String(mmetric.AttrHTTPRequestMethod, r.Request.Method),
		attribute.String(mmetric.AttrNetworkProtocolVersion, protocolVersion),
	}
	requestOptions := mmetric.WithAttributes(requestAttributes...)

	responseAttributes := make([]attribute.KeyValue, 0, len(requestAttributes)+2)
	responseAttributes = append(responseAttributes, requestAttributes...)
	responseAttributes = append(responseAttributes, attribute.Int(mmetric.AttrHTTPResponseStatusCode, r.Writer.Status()))

	if len(r.Errors) > 0 {
		var errCode int
		if err := r.Errors[0]; err != nil {
			errCode = merror.Code(err).Code()
		}
		responseAttributes = append(responseAttributes, attribute.Int(mmetric.AttrErrorCode, errCode))
	}
	responseOptions := mmetric.WithAttributes(responseAttributes...)

	// increase request total
	metricManager.HTTPServerRequestTotal.Inc(ctx, responseOptions)

	// decrease active request count
	metricManager.HTTPServerRequestActive.Dec(ctx, requestOptions)

	// record response body size
	metricManager.HTTPServerResponseBodySize.Add(
		ctx,
		float64(r.Writer.Size()),
		responseOptions,
	)

	// record request duration total
	metricManager.HTTPServerRequestDurationTotal.Add(
		ctx,
		durationMilli,
		responseOptions,
	)

	// record request duration distribution
	metricManager.HTTPServerRequestDuration.Record(
		durationMilli,
		responseOptions,
	)
}
