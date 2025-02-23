package provider

import (
	"context"
	"encoding/binary"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// IDGenerator 实现了 trace.IDGenerator 接口
type IDGenerator struct {
	spanCounter uint64
}

// NewIDGenerator 返回一个新的 IDGenerator 实例
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// NewIDs 生成新的 trace 和 span ID
func (gen *IDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	return gen.newTraceID(), gen.newSpanID()
}

// NewSpanID 生成一个新的 span ID
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
