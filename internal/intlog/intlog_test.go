package intlog

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestLoggingHonorsDebugFlag(t *testing.T) {
	SetDebug(false)
	assert.Empty(t, captureOutput(t, func() {
		Print(context.Background(), "hidden")
		Printf(context.Background(), "hidden: %d", 1)
		Error(context.Background(), "hidden")
		Errorf(context.Background(), "hidden: %d", 1)
	}))

	SetDebug(true)
	t.Cleanup(func() { SetDebug(false) })
	output := captureOutput(t, func() {
		Print(context.Background(), "plain")
		Printf(context.Background(), "formatted: %d", 2)
		Error(context.Background(), "failure")
		Errorf(context.Background(), "failure: %d", 3)
	})

	assert.Contains(t, output, "[INTE]")
	assert.Contains(t, output, "plain")
	assert.Contains(t, output, "formatted: 2")
	assert.Contains(t, output, "failure")
	assert.Contains(t, output, "Caller Stack:")
	assert.Contains(t, output, "intlog_test.go")
}

func TestTraceIDAndStackHelpers(t *testing.T) {
	assert.Empty(t, traceIDStr(nil))
	assert.Empty(t, traceIDStr(context.Background()))

	traceID := trace.TraceID{1, 2, 3, 4}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)
	assert.Equal(t, "{"+traceID.String()+"}", traceIDStr(ctx))

	assert.True(t, isFilteredStack("/tmp/internal/intlog/intlog.go"))
	assert.False(t, isFilteredStack("/tmp/service.go"))
	assert.NotEmpty(t, getCallerStack())
	assert.Contains(t, file(), ":")

	SetDebug(true)
	t.Cleanup(func() { SetDebug(false) })
	assert.Contains(t, captureOutput(t, func() { Print(ctx, "traced") }), traceID.String())
}

func captureOutput(t *testing.T, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	original := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = original })

	fn()
	require.NoError(t, writer.Close())
	os.Stdout = original

	var output bytes.Buffer
	_, err = io.Copy(&output, reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	return output.String()
}
