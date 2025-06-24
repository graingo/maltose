package mlog

// SetConfig sets the logger configuration.
func SetConfig(config *Config) error {
	return defaultLogger.SetConfig(config)
}

// SetPath sets the log file path.
func SetPath(path string) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"path": path,
	})
}

// SetTimeFormat sets the log time format.
func SetTimeFormat(timeFormat string) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"time_format": timeFormat,
	})
}

// SetFormat sets the log format.
func SetFormat(format string) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"format": format,
	})
}

// SetStdoutPrint sets the stdout print.
func SetStdoutPrint(enabled bool) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"stdout": enabled,
	})
}

// SetFile sets the log file name, supporting date patterns.
func SetFile(file string) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"file": file,
	})
}

// SetAutoClean sets the number of days to keep log files.
func SetAutoClean(days int) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"auto_clean": days,
	})
}

// SetCtxKeys sets the context keys to extract values from.
func SetCtxKeys(keys []string) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"ctx_keys": keys,
	})
}
