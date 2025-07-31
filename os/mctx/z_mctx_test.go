package mctx_test

import (
	"context"
	"testing"

	"github.com/graingo/maltose/os/mctx"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestMain(m *testing.M) {
	// a silent tracer provider for testing
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	m.Run()
}

func TestNew(t *testing.T) {
	ctx := mctx.New()
	assert.NotNil(t, ctx)
	assert.NotEmpty(t, mctx.CtxID(ctx))
}

func TestWithSpan(t *testing.T) {
	t.Run("with_background_context", func(t *testing.T) {
		ctx := mctx.WithSpan(context.Background(), "test_span")
		assert.NotNil(t, ctx)
		assert.NotEmpty(t, mctx.CtxID(ctx))
	})

	t.Run("with_existing_context_id", func(t *testing.T) {
		ctx1 := mctx.New()
		ctx1ID := mctx.CtxID(ctx1)
		assert.NotEmpty(t, ctx1ID)

		ctx2 := mctx.WithSpan(ctx1, "another_span")
		ctx2ID := mctx.CtxID(ctx2)

		assert.NotEqual(t, ctx1, ctx2)
		assert.Equal(t, ctx1ID, ctx2ID)
	})

	t.Run("with_empty_span_name", func(t *testing.T) {
		ctx := mctx.WithSpan(context.Background(), "")
		assert.NotNil(t, ctx)
		assert.NotEmpty(t, mctx.CtxID(ctx))
	})
}

func TestCtxID(t *testing.T) {
	t.Run("with_id", func(t *testing.T) {
		ctx := mctx.New()
		assert.NotEmpty(t, mctx.CtxID(ctx))
	})
	t.Run("without_id", func(t *testing.T) {
		assert.Empty(t, mctx.CtxID(context.Background()))
	})
}
