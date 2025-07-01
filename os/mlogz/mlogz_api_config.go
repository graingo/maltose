package mlogz

// SetConfig sets the logger configuration.
func SetConfig(config *Config) error {
	return defaultLogger.SetConfig(config)
}

// SetFilepath sets the log file path.
func SetFilepath(path string) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"filepath": path,
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

// SetStdout sets the stdout print.
func SetStdout(enabled bool) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"stdout": enabled,
	})
}

// SetMaxSize sets the max size of the log file.
func SetMaxSize(maxSize int) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"max_size": maxSize,
	})
}

// SetMaxBackups sets the max backups of the log file.
func SetMaxBackups(maxBackups int) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"max_backups": maxBackups,
	})
}

// SetMaxAge sets the max age of the log file.
func SetMaxAge(maxAge int) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"max_age": maxAge,
	})
}

// SetCtxKeys sets the context keys to extract values from.
func SetCtxKeys(keys []string) {
	defaultLogger.SetConfigWithMap(map[string]any{
		"ctx_keys": keys,
	})
}

// SetLevel sets the log level.
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// GetLevel returns the log level.
func GetLevel() Level {
	return defaultLogger.GetLevel()
}
