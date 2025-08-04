package mlog_test

import (
	"context"
	"path/filepath"
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
	tempDir := b.TempDir()
	logPath := filepath.Join(tempDir, "benchmark.log")
	// Redirecting output to /dev/null is a standard technique for benchmarking I/O components
	// without measuring the actual I/O performance. This is necessary because mlog does not
	// currently provide a public API to set the output writer directly to io.Discard.
	cfg := mlog.Config{
		Level:    mlog.DebugLevel,
		Filepath: logPath, // Write to the null device to discard output.
		Stdout:   false,
		Format:   "json",
	}

	logger := mlog.New(&cfg)
	b.Cleanup(func() {
		_ = logger.Close()
	})
	return logger
}

func BenchmarkLogger(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		logger := setupBenchmarkLogger(b)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(ctx, "a simple message")
			}
		})
	})

	b.Run("sprintf", func(b *testing.B) {
		logger := setupBenchmarkLogger(b)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infof(ctx, "a message with formatting: %s %d", "hello", 123)
			}
		})
	})

	b.Run("with_1_field", func(b *testing.B) {
		logger := setupBenchmarkLogger(b)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(ctx, "message with one field", oneMlogField...)
			}
		})
	})

	b.Run("with_5_fields", func(b *testing.B) {
		logger := setupBenchmarkLogger(b)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(ctx, "message with five fields", fiveMlogFields...)
			}
		})
	})

	b.Run("with_10_fields", func(b *testing.B) {
		logger := setupBenchmarkLogger(b)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(ctx, "message with ten fields", tenMlogFields...)
			}
		})
	})

	b.Run("with_logger_10_fields", func(b *testing.B) {
		logger := setupBenchmarkLogger(b)
		withLogger := logger.With(tenMlogFields...)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				withLogger.Infow(ctx, "message from a contextual logger")
			}
		})
	})

	b.Run("with_hooks", func(b *testing.B) {
		logger := setupBenchmarkLogger(b)
		logger.AddHook(&benchmarkHook{})

		type ctxKey string
		const ctxKeyRequestID ctxKey = "request_id"
		requestCtx := context.WithValue(context.Background(), ctxKeyRequestID, "req-12345")

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Infow(requestCtx, "message that will trigger the hook")
			}
		})
	})
}

type benchmarkHook struct{}

func (h *benchmarkHook) Name() string { return "benchmark_hook" }

func (h *benchmarkHook) Levels() []mlog.Level { return mlog.AllLevels() }

func (h *benchmarkHook) Fire(entry *mlog.Entry) {
	if value := entry.GetContext().Value("request_id"); value != nil {
		if str, ok := value.(string); ok {
			entry.AddField(mlog.String("request_id", str))
		}
	}
}
