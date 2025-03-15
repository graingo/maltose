package provider

import (
	"context"
	"encoding/binary"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// IDGenerator implements the trace.IDGenerator interface.
type IDGenerator struct {
	spanCounter uint64
}

// NewIDGenerator returns a new IDGenerator instance.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// NewIDs generates new trace and span IDs.
func (gen *IDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	return gen.newTraceID(), gen.newSpanID()
}

// NewSpanID generates a new span ID.
func (gen *IDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	return gen.newSpanID()
}

func (gen *IDGenerator) newTraceID() trace.TraceID {
	tid := [16]byte{}
	now := time.Now().UnixNano()
	binary.BigEndian.PutUint64(tid[0:8], uint64(now))
	binary.BigEndian.PutUint64(tid[8:16], uint64(time.Now().UnixNano()))
	return trace.TraceID(tid)
}

func (gen *IDGenerator) newSpanID() trace.SpanID {
	sid := [8]byte{}
	binary.BigEndian.PutUint64(sid[:], atomic.AddUint64(&gen.spanCounter, 1))
	return trace.SpanID(sid)
}
