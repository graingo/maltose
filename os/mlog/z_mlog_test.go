package mlog_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"io"
	"os/exec"

	"github.com/graingo/maltose/errors/merror"
	"github.com/graingo/maltose/net/mtrace"
	"github.com/graingo/maltose/os/mctx"
	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestLogger creates a logger that writes to a temporary file,
// returning the logger instance and the log file path.
// It automatically registers a cleanup function to close the logger after the test.
func setupTestLogger(t *testing.T, cfg mlog.Config) (*mlog.Logger, string) {
	t.Helper()
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Set default values for testing to ensure logs are written to our specified file.
	if cfg.Level.String() == "unknown" { // Check for the default zero value.
		cfg.Level = mlog.DebugLevel
	}
	cfg.Filepath = logPath
	cfg.Stdout = false // Disable stdout in tests to avoid noisy output.
	if cfg.Format == "" {
		cfg.Format = "json" // Default to json only if not specified.
	}

	logger := mlog.New(&cfg)

	// Use t.Cleanup to register a deferred function that will execute after each test (or subtest).
	// This ensures that resources are correctly released even if the test fails.
	t.Cleanup(func() {
		// A small delay to ensure log rotation goroutine has time to process if needed.
		// time.Sleep(10 * time.Millisecond) // No longer needed as we close manually
		// err := logger.Close() // Removed to prevent double-closing. Tests that need flushing will call it manually.
		// assert.NoError(t, err, "logger should be closed without error")
	})

	return logger, logPath
}

// setupDefaultLogger configures the package-level default logger for testing.
func setupDefaultLogger(t *testing.T, cfg mlog.Config) string {
	t.Helper()
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "api_test.log")

	// Apply test-specific defaults.
	cfg.Filepath = logPath
	cfg.Stdout = false
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	if cfg.Level.String() == "unknown" {
		cfg.Level = mlog.DebugLevel
	}

	// Reset the default logger to a clean state for each test.
	err := mlog.SetConfig(&cfg)
	require.NoError(t, err, "failed to configure default logger for API test")

	t.Cleanup(func() {
		// The test function is now responsible for calling Close.
		// We can add a safeguard here if needed, but for now, we rely on manual closing.
	})

	return logPath
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

// --- Logger Instance Tests ---

func TestLoggerInstance(t *testing.T) {
	t.Run("basic_output", func(t *testing.T) {
		logger, logPath := setupTestLogger(t, mlog.Config{})

		logger.Infow(context.Background(), "user logged in", mlog.String("username", "test"), mlog.Int("user_id", 123))

		time.Sleep(10 * time.Millisecond)

		logStr := readLogFile(t, logPath)
		assert.Contains(t, logStr, `"level":"info"`)
		assert.Contains(t, logStr, `"msg":"user logged in"`)
		assert.Contains(t, logStr, `"username":"test"`)
		assert.Contains(t, logStr, `"user_id":123`)
	})

	t.Run("dynamic_level_change", func(t *testing.T) {
		logger, logPath := setupTestLogger(t, mlog.Config{Level: mlog.DebugLevel})

		logger.Debugf(context.Background(), "this is a debug message")
		logger.SetLevel(mlog.InfoLevel)
		logger.Debugf(context.Background(), "this should not be logged")
		logger.Infof(context.Background(), "this is an info message")

		time.Sleep(10 * time.Millisecond)

		logContent := readLogFile(t, logPath)
		assert.Contains(t, logContent, "this is a debug message")
		assert.NotContains(t, logContent, "this should not be logged")
		assert.Contains(t, logContent, "this is an info message")
	})

	t.Run("trace_and_ctx_keys_hooks", func(t *testing.T) {
		const ctxKeyRequestID mlog.CtxKey = "request_id"
		logger, logPath := setupTestLogger(t, mlog.Config{
			CtxKeys: []string{"request_id"},
		})

		ctx := mctx.New()
		ctx, _ = mtrace.WithTraceID(ctx, "12345678901234567890123456789012")
		ctx = context.WithValue(ctx, ctxKeyRequestID, "req-abcde")

		logger.Infow(ctx, "testing hooks")
		time.Sleep(10 * time.Millisecond)

		logStr := readLogFile(t, logPath)
		var logMap map[string]interface{}
		for _, line := range strings.Split(strings.TrimSpace(logStr), "\n") {
			if strings.Contains(line, "testing hooks") {
				err := json.Unmarshal([]byte(line), &logMap)
				require.NoError(t, err, "log output line should be valid json")
				assert.Equal(t, "12345678901234567890123456789012", logMap["trace.id"])
				assert.Equal(t, "req-abcde", logMap["request_id"])
				return
			}
		}
		t.Fatal("did not find the expected log line with hook attributes")
	})

	t.Run("with_adds_persistent_fields", func(t *testing.T) {
		baseLogger, logPath := setupTestLogger(t, mlog.Config{})

		serviceLogger := baseLogger.With(
			mlog.String("service", "payment-service"),
			mlog.String("version", "1.2.3"),
		)

		serviceLogger.Infof(context.Background(), "processing payment")
		serviceLogger.Errorw(context.Background(), nil, "payment failed", mlog.Int("order_id", 456))

		time.Sleep(10 * time.Millisecond)

		logStr := readLogFile(t, logPath)
		lines := strings.Split(strings.TrimSpace(logStr), "\n")
		require.Len(t, lines, 2, "should have two log lines")

		assert.Contains(t, lines[0], `"service":"payment-service"`)
		assert.Contains(t, lines[0], `"version":"1.2.3"`)
		assert.Contains(t, lines[0], `"msg":"processing payment"`)

		assert.Contains(t, lines[1], `"service":"payment-service"`)
		assert.Contains(t, lines[1], `"version":"1.2.3"`)
		assert.Contains(t, lines[1], `"order_id":456`)
	})

	t.Run("add_and_fire_custom_hook", func(t *testing.T) {
		logger, logPath := setupTestLogger(t, mlog.Config{})
		customHook := &customHookForTest{AppName: "my-test-app"}
		logger.AddHook(customHook)

		logger.Warnf(context.Background(), "this is a warning with a custom hook")
		time.Sleep(10 * time.Millisecond)

		logStr := readLogFile(t, logPath)
		var logMap map[string]interface{}

		require.True(t, strings.Contains(logStr, "custom hook"), "log should contain the message")
		err := json.Unmarshal([]byte(logStr), &logMap)
		require.NoError(t, err, "log output should be valid json")

		assert.Equal(t, "my-test-app", logMap["app_name"])
		assert.Equal(t, "warn", logMap["level"])
	})

	t.Run("remove_hook", func(t *testing.T) {
		logger, logPath := setupTestLogger(t, mlog.Config{})
		customHook := &customHookForTest{AppName: "my-removable-app"}
		logger.AddHook(customHook)

		logger.Infof(context.Background(), "message with hook")
		logger.RemoveHook(customHook.Name())
		logger.Infof(context.Background(), "message without hook")

		time.Sleep(10 * time.Millisecond)

		logStr := readLogFile(t, logPath)
		lines := strings.Split(strings.TrimSpace(logStr), "\n")
		require.Len(t, lines, 2, "should have two log lines")

		var logMap1 map[string]interface{}
		err := json.Unmarshal([]byte(lines[0]), &logMap1)
		require.NoError(t, err)
		assert.Equal(t, "my-removable-app", logMap1["app_name"])

		var logMap2 map[string]interface{}
		err = json.Unmarshal([]byte(lines[1]), &logMap2)
		require.NoError(t, err)
		assert.Nil(t, logMap2["app_name"], "app_name field should not exist after hook is removed")
	})
}

// customHookForTest is a simple hook for testing purposes that adds a static field.
type customHookForTest struct {
	AppName string
}

func (h *customHookForTest) Name() string {
	return "custom_app_hook"
}

func (h *customHookForTest) Levels() []mlog.Level {
	// Apply to all log levels.
	return mlog.AllLevels()
}

func (h *customHookForTest) Fire(entry *mlog.Entry) {
	// Add a static app_name field to every log entry.
	entry.AddField(mlog.String("app_name", h.AppName))
}

// --- Package-Level API Tests ---

func TestPackageAPI(t *testing.T) {
	t.Run("api_functions", func(t *testing.T) {
		logPath := setupDefaultLogger(t, mlog.Config{})

		mlog.Infof(context.Background(), "API test: %s", "hello")
		mlog.Warnw(context.Background(), "API warning", mlog.String("reason", "test-case"))
		testErr := merror.New("this is a test error")
		mlog.Errorw(context.Background(), testErr, "API error occurred")

		logStr := readLogFile(t, logPath)
		lines := strings.Split(strings.TrimSpace(logStr), "\n")
		require.GreaterOrEqual(t, len(lines), 3, "expected at least 3 log lines from API calls")

		assert.Contains(t, lines[0], `"level":"info"`)
		assert.Contains(t, lines[0], `"msg":"API test: hello"`)

		assert.Contains(t, lines[1], `"level":"warn"`)
		assert.Contains(t, lines[1], `"msg":"API warning"`)
		assert.Contains(t, lines[1], `"reason":"test-case"`)

		assert.Contains(t, lines[2], `"level":"error"`)
		assert.Contains(t, lines[2], `"msg":"API error occurred"`)
		assert.Contains(t, lines[2], `"error":"this is a test error"`)
		mlog.Close() // Manually close to flush
	})

	t.Run("api_level_change", func(t *testing.T) {
		logPath := setupDefaultLogger(t, mlog.Config{Level: mlog.InfoLevel})

		mlog.Debugf(context.Background(), "you can't see me")
		mlog.Infof(context.Background(), "initial message")

		mlog.SetLevel(mlog.DebugLevel)
		assert.Equal(t, mlog.DebugLevel, mlog.GetLevel())

		mlog.Debugf(context.Background(), "now you can see me")
		mlog.Close() // Manually close to flush

		logStr := readLogFile(t, logPath)
		assert.NotContains(t, logStr, "you can't see me")
		assert.Contains(t, logStr, "initial message")
		assert.Contains(t, logStr, "now you can see me")
	})

	t.Run("api_global_config", func(t *testing.T) {
		logPath := setupDefaultLogger(t, mlog.Config{
			Level:    mlog.InfoLevel,
			Caller:   true,
			Filepath: "global_config.log", // provide a name for clarity
		})
		defer mlog.Close()

		mlog.Infof(context.Background(), "global config info")

		content, err := os.ReadFile(logPath)
		require.NoError(t, err)

		logStr := string(content)
		// The caller will now be the API file, not the test file, because callerSkip is fixed.
		assert.Contains(t, logStr, "mlog_api.go")
		assert.NotContains(t, logStr, "z_mlog_test.go")
		assert.Contains(t, logStr, "global config info")
	})

	t.Run("api_kitchen_sink", func(t *testing.T) {
		// This test calls many of the previously uncovered log functions.
		logPath := setupDefaultLogger(t, mlog.Config{Level: mlog.DebugLevel})

		ctx := context.Background()
		mlog.Debugw(ctx, "debugw message", mlog.String("field1", "value1"))
		mlog.Infow(ctx, "infow message", mlog.Int("field2", 123))
		mlog.Warnf(ctx, "warnf message: %s", "formatted")
		mlog.Errorf(ctx, nil, "errorf message")

		// Test With, AddHook, RemoveHook on the default logger
		hook := &customHookForTest{AppName: "global-app"}
		loggerWith := mlog.With(mlog.String("component", "global"))
		loggerWith.AddHook(hook)
		loggerWith.Errorw(ctx, nil, "with global logger")
		loggerWith.RemoveHook(hook.Name())

		// Test various field types
		mlog.Debugw(ctx, "field types",
			mlog.Duration("duration", time.Second),
			mlog.Int64("int64", 12345),
			mlog.Time("time", time.Now()),
			mlog.Uint("uint", 1),
			mlog.Uint64("uint64", 2),
		)

		mlog.Close()
		logStr := readLogFile(t, logPath)
		assert.Contains(t, logStr, "debugw message")
		assert.Contains(t, logStr, "infow message")
		assert.Contains(t, logStr, "warnf message: formatted")
		assert.Contains(t, logStr, "errorf message")
		assert.Contains(t, logStr, "with global logger")
		assert.Contains(t, logStr, `"app_name":"global-app"`)
		assert.Contains(t, logStr, `"component":"global"`)
		assert.Contains(t, logStr, "field types")
		assert.Contains(t, logStr, "duration")
	})
}

// --- File Rotation Tests ---

func TestLogger_FileRotation(t *testing.T) {
	t.Run("rotation_by_size", func(t *testing.T) {
		tempDir := t.TempDir()
		logPath := filepath.Join(tempDir, "rotation_size.log")

		// Set a small rotation size to trigger it easily.
		logger := mlog.New(&mlog.Config{
			Filepath: logPath,
			Stdout:   false,
			Level:    mlog.DebugLevel,
			Format:   "text",
			MaxSize:  1, // 1 MB
		})

		// Write more than 1MB of data to trigger rotation.
		longLogLine := strings.Repeat("a", 1024) // 1KB string
		totalSize := 0
		for i := 0; i < 1025; i++ {
			line := longLogLine + "\n"
			totalSize += len(line)
			logger.Infof(context.Background(), line)
		}

		// Closing the logger ensures all buffered writes are flushed.
		require.NoError(t, logger.Close())

		// After rotation, the current log file size should be much smaller than the total size written.
		fileInfo, err := os.Stat(logPath)
		require.NoError(t, err)
		assert.Less(t, fileInfo.Size(), int64(totalSize), "current log file should be smaller than total written data after rotation")

		// Also, check that at least one backup file was created.
		// The pattern needs to match `rotation_size.<timestamp>.log`
		globPattern := filepath.Join(tempDir, "rotation_size.*.log")
		files, err := filepath.Glob(globPattern)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(files), 1, "should have at least one rotated log file")
	})

	t.Run("rotation_backup_limit", func(t *testing.T) {
		tempDir := t.TempDir()
		logPath := filepath.ToSlash(filepath.Join(tempDir, "rotation_backup.log"))

		logger := mlog.New(&mlog.Config{
			Filepath:   logPath,
			Stdout:     false,
			Level:      mlog.DebugLevel,
			Format:     "text",
			MaxSize:    1, // 1 MB
			MaxBackups: 2, // Keep only 2 backup files
		})

		// Write enough data to trigger multiple rotations.
		longLogLine := strings.Repeat("a", 1024*512) // 0.5MB
		for i := 0; i < 5; i++ {                     // Write ~2.5MB to trigger 2 rotations
			logger.Infof(context.Background(), longLogLine)
		}

		require.NoError(t, logger.Close())

		files, err := filepath.Glob(logPath + ".*")
		require.NoError(t, err)

		// The number of backup files should not exceed the limit.
		assert.LessOrEqual(t, len(files), 2, "number of backup files should not exceed MaxBackups")
	})

	/*
		t.Run("rotation_by_age", func(t *testing.T) {
			tempDir := t.TempDir()
			logPath := filepath.ToSlash(filepath.Join(tempDir, "rotation_age.log"))

			// Create some fake old log files to be cleaned up.
			oldTime := time.Now().Add(-3 * 24 * time.Hour) // 3 days ago
			fakeOldLog := logPath + "." + oldTime.Format("2006-01-02T15-04-05.000")
			err := os.WriteFile(fakeOldLog, []byte("old log"), 0644)
			require.NoError(t, err)

			logger := mlog.New(&mlog.Config{
				Filepath: logPath,
				Stdout:   false,
				Level:    mlog.DebugLevel,
				MaxSize:  1, // 1 MB to ensure rotation check triggers cleanup
				MaxAge:   1, // Keep logs for only 1 day
			})

			// A single write is enough to trigger the cleanup check.
			logger.Infof(context.Background(), "trigger cleanup")
			require.NoError(t, logger.Close())

			// The old log file should have been removed.
			_, err = os.Stat(fakeOldLog)
			assert.True(t, os.IsNotExist(err), "old log file should be removed by MaxAge rule")
		})
	*/
}

// --- Output, Levels, and Error Handling ---

func TestLogger_OutputAndErrors(t *testing.T) {
	t.Run("text_format_output", func(t *testing.T) {
		logger, logPath := setupTestLogger(t, mlog.Config{Format: "text"})
		logger.Infow(context.Background(), "text message", mlog.String("key", "value"))
		require.NoError(t, logger.Close())

		logStr := readLogFile(t, logPath)
		// Zap's ConsoleEncoder format is like: TIME\tLEVEL\tMESSAGE\t{"key": "value"}
		assert.NotContains(t, logStr, `"level":`)
		assert.NotContains(t, logStr, `"msg":`)
		assert.Contains(t, logStr, "\tinfo\t")
		assert.Contains(t, logStr, "text message")
		assert.Contains(t, logStr, `{"key": "value"}`) // Corrected assertion for zap's JSON with spaces
	})

	t.Run("log_levels_are_respected", func(t *testing.T) {
		logger, logPath := setupTestLogger(t, mlog.Config{Level: mlog.WarnLevel})
		logger.Infof(context.Background(), "this should not be logged")
		logger.Warnf(context.Background(), "this should be logged")
		require.NoError(t, logger.Close())

		logStr := readLogFile(t, logPath)
		assert.NotContains(t, logStr, "this should not be logged")
		assert.Contains(t, logStr, "this should be logged")
	})

	t.Run("invalid_log_path_does_not_write", func(t *testing.T) {
		// Create a read-only directory to simulate a permission error.
		tempDir := t.TempDir()
		readOnlyDir := filepath.Join(tempDir, "readonly_dir")
		err := os.Mkdir(readOnlyDir, 0555) // Read and execute permissions, but not write.
		require.NoError(t, err)

		invalidPath := filepath.Join(readOnlyDir, "test.log")

		// New() itself might not fail.
		logger := mlog.New(&mlog.Config{
			Filepath: invalidPath,
			Stdout:   false,
		})
		require.NotNil(t, logger)

		// The error might be silent, so we check if the file was created.
		// It should not have been.
		err = logger.Close()
		assert.NoError(t, err, "close should not return an error on silent write failure")

		// Verify that no log file was created in the read-only directory.
		_, err = os.Stat(invalidPath)
		assert.True(t, os.IsNotExist(err), "log file should not be created in a read-only directory")
	})

	t.Run("set_config_with_map_invalid_level", func(t *testing.T) {
		cfg := mlog.Config{}
		err := cfg.SetConfigWithMap(map[string]interface{}{
			"level": "invalid-level-string",
		})
		assert.Error(t, err)
	})
}

// --- Termination Tests ---

func TestLogger_Termination(t *testing.T) {
	// This helper function runs the test in a subprocess to safely test os.Exit calls.
	runInSubprocess := func(t *testing.T, testName string) (string, error) {
		// Re-run the test binary, but only for the specific sub-test we want.
		cmd := exec.Command(os.Args[0], "-test.run", "^"+t.Name()+"$/^"+testName+"$")
		// Set an environment variable to signal that this is the subprocess execution.
		cmd.Env = append(os.Environ(), "MLOG_SUBPROCESS_TEST=1")
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	t.Run("fatal_level_exits", func(t *testing.T) {
		// This part of the test runs in the main process.
		if os.Getenv("MLOG_SUBPROCESS_TEST") != "1" {
			_, err := runInSubprocess(t, t.Name())
			// A call to Fatal should result in a non-zero exit code.
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok, "expected an exit error from subprocess")
			assert.False(t, exitErr.Success(), "process should exit with non-zero status")
		} else {
			// This part runs in the subprocess.
			// Configure a logger that won't write to a file to keep things simple.
			logger := mlog.New(&mlog.Config{Stdout: false, Writer: io.Discard})
			logger.Fatalf(context.Background(), nil, "fatal exit test") // This will call os.Exit(1)
		}
	})

	t.Run("panic_level_panics", func(t *testing.T) {
		// This part of the test runs in the main process.
		if os.Getenv("MLOG_SUBPROCESS_TEST") != "1" {
			output, err := runInSubprocess(t, t.Name())
			// A panic should result in a non-zero exit code and stderr output containing the panic message.
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok, "expected an exit error from subprocess")
			assert.False(t, exitErr.Success(), "process should exit with non-zero status")
			assert.Contains(t, string(output), "panic: panic exit test")
		} else {
			// This part runs in the subprocess.
			logger := mlog.New(&mlog.Config{Stdout: false, Writer: io.Discard})
			logger.Panicf(context.Background(), nil, "panic exit test") // This will panic
		}
	})
}

// --- Advanced Feature Tests ---

func TestLogger_AdvancedFeatures(t *testing.T) {
	t.Run("named_instances", func(t *testing.T) {
		// Get two named loggers. They should be different instances.
		logger1 := mlog.Instance("metrics")
		logger2 := mlog.Instance("events")
		assert.NotSame(t, logger1, logger2)

		// Retrieving the same name should return the same instance.
		logger1_again := mlog.Instance("metrics")
		assert.Same(t, logger1, logger1_again)

		// The default instance is different from named instances.
		defaultLogger := mlog.Instance()
		assert.NotSame(t, defaultLogger, logger1)
	})

	t.Run("set_config_with_map_retains_state", func(t *testing.T) {
		// 1. Setup logger with initial state: caller=true.
		logger, logPath := setupTestLogger(t, mlog.Config{
			Caller: true,
		})

		// 2. Add a persistent field using With().
		loggerWith := logger.With(mlog.String("component", "database"))

		// 3. Use SetConfigWithMap to change the level.
		// This should not affect the 'caller' setting or the 'With' fields.
		err := loggerWith.SetConfigWithMap(map[string]interface{}{
			"level": "warn",
		})
		require.NoError(t, err)

		// 4. Log messages and close to flush.
		loggerWith.Infof(context.Background(), "should not be logged")
		loggerWith.Warnf(context.Background(), "should be logged and have state")
		require.NoError(t, loggerWith.Close())

		// 5. Assertions
		logStr := readLogFile(t, logPath)

		// a. Assert level change was successful.
		assert.NotContains(t, logStr, "should not be logged")
		assert.Contains(t, logStr, "should be logged and have state")

		// b. Assert the 'With' field was retained.
		assert.Contains(t, logStr, `"component":"database"`)

		// c. Assert the 'caller' setting was retained.
		// The caller of loggerWith.Warnf is this test file.
		assert.Contains(t, logStr, "z_mlog_test.go")
	})

	t.Run("parse_invalid_level", func(t *testing.T) {
		_, err := mlog.ParseLevel("nonexistent_level")
		assert.Error(t, err)
	})
}
