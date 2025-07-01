package mlog_test

import (
	"context"
	"testing"

	"github.com/graingo/maltose/os/mlog"
)

var (
	// A static context to avoid context creation overhead in benchmarks.
	ctx = context.Background()
	// Pre-allocate fields to avoid allocation overhead inside the benchmark loop itself.
	tenMlogFields = []mlog.Field{
		mlog.String("key1", "value1"),
		mlog.Int("key2", 123),
		mlog.Bool("key3", true),
		mlog.Float64("key4", 123.456),
		mlog.String("key5", "value5"),
		mlog.String("key6", "value6"),
		mlog.String("key7", "value7"),
		mlog.String("key8", "value8"),
		mlog.String("key9", "value9"),
		mlog.String("key10", "value10"),
	}
	oneMlogField   = tenMlogFields[:1]
	fiveMlogFields = tenMlogFields[:5]
)

// setupBenchmarkLogger creates a logger that writes to a null device,
// effectively isolating the benchmark to the logger's processing overhead by eliminating I/O.
func setupBenchmarkLogger(b *testing.B) *mlog.Logger {
	b.Helper()
	// Redirecting output to /dev/null is a standard technique for benchmarking I/O components
	// without measuring the actual I/O performance. This is necessary because mlog does not
	// currently provide a public API to set the output writer directly to io.Discard.
	cfg := mlog.Config{
		Level:    mlog.DebugLevel,
		Filepath: "", // Write to the null device to discard output.
		Stdout:   false,
		Format:   "json",
	}

	logger := mlog.New(&cfg)
	b.Cleanup(func() {
		_ = logger.Close()
	})
	return logger
}

func BenchmarkMlog_Simple(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "a simple message")
		}
	})
}

func BenchmarkMlog_Sprintf(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof(ctx, "a message with formatting: %s %d", "hello", 123)
		}
	})
}

func BenchmarkMlog_With1Field(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "message with one field", oneMlogField...)
		}
	})
}

func BenchmarkMlog_With5Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "message with five fields", fiveMlogFields...)
		}
	})
}

func BenchmarkMlog_With10Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "message with ten fields", tenMlogFields...)
		}
	})
}

// BenchmarkMlog_WithLogger10Fields tests the performance of a logger that has been pre-configured
// with contextual fields using the With() method. This is a very common and important use case.
func BenchmarkMlog_WithLogger10Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	withLogger := logger.With(tenMlogFields...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			withLogger.Infow(ctx, "message from a contextual logger")
		}
	})
}
