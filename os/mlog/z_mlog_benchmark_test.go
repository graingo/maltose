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
	tenLogrusFields = mlog.Fields{
		"key1":  "value1",
		"key2":  123,
		"key3":  true,
		"key4":  123.456,
		"key5":  "value5",
		"key6":  "value6",
		"key7":  "value7",
		"key8":  "value8",
		"key9":  "value9",
		"key10": "value10",
	}
	fiveLogrusFields mlog.Fields
)

func init() {
	// Helper to create a 5-field map from the 10-field map.
	fiveLogrusFields = make(mlog.Fields, 5)
	i := 0
	for k, v := range tenLogrusFields {
		if i >= 5 {
			break
		}
		fiveLogrusFields[k] = v
		i++
	}
}

// setupBenchmarkLogger creates a logger that writes to io.Discard,
// effectively isolating the benchmark to the logger's processing overhead by eliminating I/O.
func setupBenchmarkLogger(b *testing.B) *mlog.Logger {
	b.Helper()
	// mlog is well-designed for benchmarking, as disabling both Stdout and Filepath
	// in the config will automatically set the output to io.Discard.
	cfg := mlog.Config{
		Level:    mlog.DebugLevel,
		Stdout:   false,
		Filepath: "",
		Format:   "json",
	}

	logger := mlog.New(&cfg)
	return logger
}

func BenchmarkMlog_Simple(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(ctx, "a simple message")
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

func BenchmarkMlog_With10Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// In mlog, WithFields creates a new logger entry on each call.
			logger.WithFields(tenLogrusFields).Info(ctx, "message with ten fields")
		}
	})
}

// BenchmarkMlog_WithLogger10Fields tests the performance of a logger that has been pre-configured
// with contextual fields using the WithFields() method. This is a very common and important use case.
func BenchmarkMlog_WithLogger10Fields(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	withLogger := logger.WithFields(tenLogrusFields)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			withLogger.Info(ctx, "message from a contextual logger")
		}
	})
}

// BenchmarkMlog_With10FieldsInArgs tests the unique mlog feature of passing fields
// directly as variadic arguments, which involves runtime reflection.
func BenchmarkMlog_With10FieldsInArgs(b *testing.B) {
	logger := setupBenchmarkLogger(b)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(ctx, "message with ten fields", tenLogrusFields)
		}
	})
}
