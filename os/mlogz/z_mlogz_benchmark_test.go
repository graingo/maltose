package mlogz_test

import (
	"context"
	"os"
	"testing"

	"github.com/graingo/maltose/os/mlogz"
)

var (
	// A static context to avoid context creation overhead in benchmarks.
	ctx = context.Background()
	// Pre-allocate fields to avoid allocation overhead inside the benchmark loop itself.
	tenMlogzFields = []mlogz.Field{
		mlogz.String("key1", "value1"),
		mlogz.Int("key2", 123),
		mlogz.Bool("key3", true),
		mlogz.Float64("key4", 123.456),
		mlogz.String("key5", "value5"),
		mlogz.String("key6", "value6"),
		mlogz.String("key7", "value7"),
		mlogz.String("key8", "value8"),
		mlogz.String("key9", "value9"),
		mlogz.String("key10", "value10"),
	}
	oneMlogzField   = tenMlogzFields[:1]
	fiveMlogzFields = tenMlogzFields[:5]
)

// setupBenchmarkLogger creates a logger that writes to a null device,
// effectively isolating the benchmark to the logger's processing overhead by eliminating I/O.
func setupBenchmarkLogger(b *testing.B) *mlogz.Logger {
	b.Helper()
	// Redirecting output to /dev/null is a standard technique for benchmarking I/O components
	// without measuring the actual I/O performance. This is necessary because mlogz does not
	// currently provide a public API to set the output writer directly to io.Discard.
	cfg := mlogz.Config{
		Level:    mlogz.DebugLevel,
		Filepath: os.DevNull, // Write to the null device to discard output.
		Stdout:   false,
		Format:   "json",
	}

	logger := mlogz.New(&cfg)
	b.Cleanup(func() {
		_ = logger.Close()
	})
	return logger
}

func BenchmarkMlogz_Simple(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "a simple message")
		}
	})
}

func BenchmarkMlogz_Sprintf(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof(ctx, "a message with formatting: %s %d", "hello", 123)
		}
	})
}

func BenchmarkMlogz_With1Field(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "message with one field", oneMlogzField...)
		}
	})
}

func BenchmarkMlogz_With5Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "message with five fields", fiveMlogzFields...)
		}
	})
}

func BenchmarkMlogz_With10Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infow(ctx, "message with ten fields", tenMlogzFields...)
		}
	})
}

// BenchmarkMlogz_WithLogger10Fields tests the performance of a logger that has been pre-configured
// with contextual fields using the With() method. This is a very common and important use case.
func BenchmarkMlogz_WithLogger10Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	withLogger := logger.With(tenMlogzFields...)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			withLogger.Infow(ctx, "message from a contextual logger")
		}
	})
}
