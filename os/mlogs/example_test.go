package mlogs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/graingo/maltose/net/mtrace"
	"github.com/graingo/maltose/os/mlogs"
)

func TestMlogs(t *testing.T) {
	ctx := context.Background()

	// 1. Basic logging
	mlogs.Info(ctx, "Server is starting...")

	// 2. Logging with attributes
	port := 8080
	mlogs.Info(ctx, "Listening on port",
		mlogs.Int("port", port),
		mlogs.String("protocol", "http"),
	)

	// 3. Error logging
	err := errors.New("something went wrong")
	mlogs.Error(ctx, err, "Failed to process request",
		mlogs.String("component", "database"),
	)

	// 4. Using With for persistent context
	requestID := "req-12345"
	reqLogger := mlogs.With(mlogs.String("request_id", requestID))
	reqLogger.Info(ctx, "Processing incoming request")
	reqLogger.Debug(ctx, "Request headers processed")

	// 5. Using With and Trace context
	ctxWithTrace, _ := mtrace.WithTraceID(ctx, "trace-abc-xyz")
	ctxWithTrace = mtrace.SetBaggageValue(ctxWithTrace, "span_id", "span-987")

	componentLogger := reqLogger.With(mlogs.String("component", "payment_service"))
	componentLogger.Warn(ctxWithTrace, "Payment processing is slow",
		mlogs.Int("duration_ms", 2500),
	)

	// 6. Formatted logging
	mlogs.Infof(ctx, "User %s logged in from %s", "mingzai", "127.0.0.1")

	// 7. Formatted error logging
	mlogs.Errorf(ctx, err, "User %s failed to login from %s", "mingzai", "127.0.0.1")
}
