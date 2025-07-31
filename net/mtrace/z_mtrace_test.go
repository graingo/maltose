package mtrace_test

import (
	"context"
	"testing"

	"github.com/graingo/maltose/net/mtrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// setupSpan creates a mock tracer and a span for testing purposes.
func setupSpan(ctx context.Context) (context.Context, trace.Span) {
	// Use the real SDK's in-memory implementation for testing instead of noop.
	tp := sdktrace.NewTracerProvider()
	tracer := tp.Tracer("test-tracer")
	return tracer.Start(ctx, "test-span")
}

func TestTraceIDAndSpanID(t *testing.T) {
	ctx, span := setupSpan(context.Background())
	spanCtx := span.SpanContext()

	t.Run("get_ids_from_valid_span", func(t *testing.T) {
		assert.Equal(t, spanCtx.TraceID().String(), mtrace.GetTraceID(ctx))
		assert.Equal(t, spanCtx.SpanID().String(), mtrace.GetSpanID(ctx))
	})

	t.Run("get_ids_from_empty_context", func(t *testing.T) {
		assert.Empty(t, mtrace.GetTraceID(context.Background()))
		assert.Empty(t, mtrace.GetSpanID(context.Background()))
	})

	t.Run("get_ids_from_nil_context", func(t *testing.T) {
		//lint:ignore SA1012 reason: testing intentional nil context handling
		assert.Empty(t, mtrace.GetTraceID(nil))
		//lint:ignore SA1012 reason: testing intentional nil context handling
		assert.Empty(t, mtrace.GetSpanID(nil))
	})
}

func TestWithTraceID(t *testing.T) {
	newTraceIDStr := "10000000000000000000000000000002"
	newTraceID, _ := trace.TraceIDFromHex(newTraceIDStr)

	t.Run("inject_into_empty_context", func(t *testing.T) {
		ctx, err := mtrace.WithTraceID(context.Background(), newTraceIDStr)
		require.NoError(t, err)
		sc := trace.SpanContextFromContext(ctx)
		assert.True(t, sc.IsValid())
		assert.Equal(t, newTraceID, sc.TraceID())
	})

	t.Run("overwrite_existing_trace_id", func(t *testing.T) {
		ctx, _ := setupSpan(context.Background())
		originalTraceID := mtrace.GetTraceID(ctx)
		require.NotEmpty(t, originalTraceID)
		require.NotEqual(t, newTraceIDStr, originalTraceID)

		ctx, err := mtrace.WithTraceID(ctx, newTraceIDStr)
		require.NoError(t, err)
		sc := trace.SpanContextFromContext(ctx)
		assert.True(t, sc.IsValid())
		assert.Equal(t, newTraceID, sc.TraceID())
	})

	t.Run("inject_invalid_trace_id", func(t *testing.T) {
		_, err := mtrace.WithTraceID(context.Background(), "invalid-trace-id")
		require.Error(t, err)
	})

	t.Run("inject_with_with_uuid", func(t *testing.T) {
		ctx, err := mtrace.WithUUID(context.Background(), newTraceIDStr)
		require.NoError(t, err)
		sc := trace.SpanContextFromContext(ctx)
		assert.True(t, sc.IsValid())
		assert.Equal(t, newTraceID, sc.TraceID())
	})
}

func TestBaggage(t *testing.T) {
	t.Run("set_and_get_single_value", func(t *testing.T) {
		ctx := context.Background()
		baggage1 := mtrace.NewBaggage(ctx)
		ctx = baggage1.SetValue("user_id", "123") // Re-assign ctx

		baggage2 := mtrace.NewBaggage(ctx)
		ctx = baggage2.SetValue("service", "auth") // Re-assign ctx

		retrievedBaggage := mtrace.NewBaggage(ctx)
		assert.Equal(t, "123", retrievedBaggage.GetVar("user_id").String())
		assert.Equal(t, "auth", retrievedBaggage.GetVar("service").String())
	})

	t.Run("set_and_get_map", func(t *testing.T) {
		ctx := context.Background()
		baggage := mtrace.NewBaggage(ctx)
		data := map[string]interface{}{
			"service": "test-service",
			"version": 1.2,
		}
		ctx = baggage.SetMap(data) // Re-assign ctx

		retrievedBaggage := mtrace.NewBaggage(ctx)
		retrievedMap := retrievedBaggage.GetMap()
		assert.Equal(t, "test-service", retrievedMap["service"])
		assert.Equal(t, "1.2", retrievedMap["version"])
	})

	t.Run("get_from_empty_context", func(t *testing.T) {
		baggage := mtrace.NewBaggage(context.Background())
		val := baggage.GetVar("non-existent")
		assert.True(t, val.IsNil())
		assert.Empty(t, baggage.GetMap())
	})

	t.Run("handle_nil_context_gracefully", func(t *testing.T) {
		//lint:ignore SA1012 reason: testing intentional nil context handling
		baggage := mtrace.NewBaggage(nil)
		ctx := baggage.SetValue("key", "value") // Re-assign ctx
		retrievedBaggage := mtrace.NewBaggage(ctx)
		assert.Equal(t, "value", retrievedBaggage.GetVar("key").String())
	})
}
