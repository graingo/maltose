package mlogz_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/graingo/maltose/net/mtrace"
	"github.com/graingo/maltose/os/mctx"
	"github.com/graingo/maltose/os/mlogz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestLogger creates a logger that writes to a temporary file,
// returning the logger instance and the log file path.
// It automatically registers a cleanup function to close the logger after the test.
func setupTestLogger(t *testing.T, cfg mlogz.Config) (*mlogz.Logger, string) {
	t.Helper()
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Set default values for testing to ensure logs are written to our specified file.
	if cfg.Level.String() == "unknown" { // Check for the default zero value.
		cfg.Level = mlogz.DebugLevel
	}
	cfg.Filepath = logPath
	cfg.Stdout = false  // Disable stdout in tests to avoid noisy output.
	cfg.Format = "json" // Explicitly set format to json for predictable test output.

	logger := mlogz.New(&cfg)

	// Use t.Cleanup to register a deferred function that will execute after each test (or subtest).
	// This ensures that resources are correctly released even if the test fails.
	t.Cleanup(func() {
		// A small delay to ensure log rotation goroutine has time to process if needed.
		time.Sleep(10 * time.Millisecond)
		err := logger.Close()
		// We don't expect Close to fail here; if it does, it indicates a serious problem.
		assert.NoError(t, err, "logger should be closed without error")
	})

	return logger, logPath
}

// readLogFile is a helper function to read the content of a log file.
func readLogFile(t *testing.T, path string) string {
	t.Helper()
	// Give a very small buffer for the OS to flush file writes.
	time.Sleep(5 * time.Millisecond)
	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read log file")
	return string(content)
}

// TestLogger_BasicOutput verifies basic log output and structured fields.
func TestLogger_BasicOutput(t *testing.T) {
	logger, logPath := setupTestLogger(t, mlogz.Config{})

	logger.Infow(context.Background(), "user logged in", mlogz.String("username", "test"), mlogz.Int("user_id", 123))

	// The t.Cleanup function will handle closing the logger.
	// We no longer call logger.Close() manually here.
	// To ensure the log is written before we read, we rely on the small delay in readLogFile.
	// For a more robust solution in a real-world scenario, one might use file system notifications.
	// But for this test, a small sleep after logging is sufficient and simpler.
	time.Sleep(10 * time.Millisecond) // Give a moment for the log to be written.

	logStr := readLogFile(t, logPath)
	assert.Contains(t, logStr, `"level":"info"`)
	assert.Contains(t, logStr, `"msg":"user logged in"`)
	assert.Contains(t, logStr, `"username":"test"`)
	assert.Contains(t, logStr, `"user_id":123`)
}

// TestLogger_DynamicLevelChange verifies dynamic log level changes.
func TestLogger_DynamicLevelChange(t *testing.T) {
	logger, logPath := setupTestLogger(t, mlogz.Config{Level: mlogz.DebugLevel})

	// 1. Initial level is Debug, so a debug message should be logged.
	logger.Debug(context.Background(), "this is a debug message")

	// 2. Raise the level to Info; subsequent debug messages should not be logged.
	logger.SetLevel(mlogz.InfoLevel)
	logger.Debug(context.Background(), "this should not be logged")
	logger.Info(context.Background(), "this is an info message")

	time.Sleep(10 * time.Millisecond)

	logContent := readLogFile(t, logPath)
	assert.Contains(t, logContent, "this is a debug message")
	assert.NotContains(t, logContent, "this should not be logged")
	assert.Contains(t, logContent, "this is an info message")
}

// TestLogger_Hooks verifies that Trace and CtxKeys hooks correctly inject fields.
func TestLogger_Hooks(t *testing.T) {
	type contextKey string
	const requestIDKey contextKey = "request_id"

	logger, logPath := setupTestLogger(t, mlogz.Config{
		CtxKeys: map[string]any{
			"request_id": requestIDKey,
		},
	})

	ctx := mctx.New()
	ctx, _ = mtrace.WithTraceID(ctx, "12345678901234567890123456789012")
	ctx = context.WithValue(ctx, requestIDKey, "req-abcde")

	logger.Infow(ctx, "testing hooks")
	time.Sleep(10 * time.Millisecond)

	logStr := readLogFile(t, logPath)
	var logMap map[string]interface{}
	// The log file may contain multiple lines; we only care about the one with our message.
	for _, line := range strings.Split(strings.TrimSpace(logStr), "\n") {
		if strings.Contains(line, "testing hooks") {
			err := json.Unmarshal([]byte(line), &logMap)
			require.NoError(t, err, "log output line should be valid json")
			assert.Equal(t, "12345678901234567890123456789012", logMap["trace_id"])
			assert.Equal(t, "req-abcde", logMap["request_id"])
			return
		}
	}
	t.Fatal("did not find the expected log line with hook attributes")
}

// TestLogger_With verifies that the With method correctly adds persistent structured fields to the logger.
func TestLogger_With(t *testing.T) {
	baseLogger, logPath := setupTestLogger(t, mlogz.Config{})

	serviceLogger := baseLogger.With(
		mlogz.String("service", "payment-service"),
		mlogz.String("version", "1.2.3"),
	)

	// First log should contain the 'With' fields.
	serviceLogger.Info(context.Background(), "processing payment")
	// Second log should also contain the 'With' fields, plus its own.
	serviceLogger.Errorw(context.Background(), nil, "payment failed", mlogz.Int("order_id", 456))

	time.Sleep(10 * time.Millisecond)

	logStr := readLogFile(t, logPath)
	lines := strings.Split(strings.TrimSpace(logStr), "\n")
	require.Len(t, lines, 2, "should have two log lines")

	// Verify the first line.
	assert.Contains(t, lines[0], `"service":"payment-service"`)
	assert.Contains(t, lines[0], `"version":"1.2.3"`)
	assert.Contains(t, lines[0], `"msg":"processing payment"`)

	// Verify the second line.
	assert.Contains(t, lines[1], `"service":"payment-service"`)
	assert.Contains(t, lines[1], `"version":"1.2.3"`)
	assert.Contains(t, lines[1], `"order_id":456`)
}

// customHookForTest is a simple hook for testing purposes that adds a static field.
type customHookForTest struct {
	AppName string
}

func (h *customHookForTest) Name() string {
	return "custom_app_hook"
}

func (h *customHookForTest) Levels() []mlogz.Level {
	// Apply to all log levels.
	return mlogz.AllLevels()
}

func (h *customHookForTest) Fire(_ context.Context, msg string, attrs []mlogz.Attr) (string, []mlogz.Attr) {
	// Add a static app_name field to every log entry.
	attrs = append(attrs, mlogz.String("app_name", h.AppName))
	return msg, attrs
}

// TestLogger_CustomHook verifies that a user-defined hook can be added and correctly fires.
func TestLogger_CustomHook(t *testing.T) {
	logger, logPath := setupTestLogger(t, mlogz.Config{})

	// Add our custom hook to the logger instance.
	customHook := &customHookForTest{AppName: "my-test-app"}
	logger.AddHook(customHook)

	// Log a message that should be processed by the hook.
	logger.Warn(context.Background(), "this is a warning with a custom hook")

	time.Sleep(10 * time.Millisecond)

	logStr := readLogFile(t, logPath)
	var logMap map[string]interface{}

	require.True(t, strings.Contains(logStr, "custom hook"), "log should contain the message")
	err := json.Unmarshal([]byte(logStr), &logMap)
	require.NoError(t, err, "log output should be valid json")

	// Verify that the field from our custom hook was added.
	assert.Equal(t, "my-test-app", logMap["app_name"])
	assert.Equal(t, "warn", logMap["level"])
}

// TestLogger_RemoveHook verifies that a hook can be correctly removed.
func TestLogger_RemoveHook(t *testing.T) {
	logger, logPath := setupTestLogger(t, mlogz.Config{})

	// Add a custom hook.
	customHook := &customHookForTest{AppName: "my-removable-app"}
	logger.AddHook(customHook)

	// Log a message, the hook should fire.
	logger.Info(context.Background(), "message with hook")

	// Remove the hook by its name.
	logger.RemoveHook(customHook.Name())

	// Log another message, the hook should NOT fire.
	logger.Info(context.Background(), "message without hook")

	time.Sleep(10 * time.Millisecond)

	logStr := readLogFile(t, logPath)
	lines := strings.Split(strings.TrimSpace(logStr), "\n")
	require.Len(t, lines, 2, "should have two log lines")

	// Verify the first log line (with the hook).
	var logMap1 map[string]interface{}
	err := json.Unmarshal([]byte(lines[0]), &logMap1)
	require.NoError(t, err)
	assert.Equal(t, "my-removable-app", logMap1["app_name"])

	// Verify the second log line (without the hook).
	var logMap2 map[string]interface{}
	err = json.Unmarshal([]byte(lines[1]), &logMap2)
	require.NoError(t, err)
	assert.Nil(t, logMap2["app_name"], "app_name field should not exist after hook is removed")
}
