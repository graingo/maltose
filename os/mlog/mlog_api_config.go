package mlog

// SetConfig sets the logger configuration.
func SetConfig(config *Config) error {
	return DefaultLogger().SetConfig(config)
}

// SetFilepath sets the log file path.
func SetFilepath(path string) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"filepath": path,
	})
}

// SetTimeFormat sets the log time format.
func SetTimeFormat(timeFormat string) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"time_format": timeFormat,
	})
}

// SetFormat sets the log format.
func SetFormat(format string) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"format": format,
	})
}

// SetStdout sets the stdout print.
func SetStdout(enabled bool) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"stdout": enabled,
	})
}

// SetMaxSize sets the max size of the log file.
func SetMaxSize(maxSize int) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"max_size": maxSize,
	})
}

// SetMaxBackups sets the max backups of the log file.
func SetMaxBackups(maxBackups int) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"max_backups": maxBackups,
	})
}

// SetMaxAge sets the max age of the log file.
func SetMaxAge(maxAge int) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"max_age": maxAge,
	})
}

// SetCtxKeys sets the context keys to extract values from.
func SetCtxKeys(keys []string) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"ctx_keys": keys,
	})
}

// SetCaller sets the caller.
func SetCaller(enabled bool) {
	_ = DefaultLogger().SetConfigWithMap(map[string]any{
		"caller": enabled,
	})
}

// SetLevel sets the log level.
func SetLevel(level Level) {
	DefaultLogger().SetLevel(level)
}

// GetLevel returns the log level.
func GetLevel() Level {
	return DefaultLogger().GetLevel()
}
