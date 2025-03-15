package mlog

// SetConfig sets the logger configuration.
func SetConfig(config Config) error {
	return defaultLogger.SetConfig(config)
}

// SetPath sets the log file path.
func SetPath(path string) {
	defaultLogger.SetPath(path)
}

// SetTimeFormat sets the log time format.
func SetTimeFormat(timeFormat string) {
	defaultLogger.SetTimeFormat(timeFormat)
}

// SetFormat sets the log format.
func SetFormat(format string) {
	defaultLogger.SetFormat(format)
}

// SetStdoutPrint sets the stdout print.
func SetStdoutPrint(enabled bool) {
	defaultLogger.SetStdoutPrint(enabled)
}

// SetFile sets the log file name, supporting date patterns.
func SetFile(file string) {
	defaultLogger.SetFile(file)
}

// SetAutoClean sets the number of days to keep log files.
func SetAutoClean(days int) {
	defaultLogger.SetAutoClean(days)
}

// SetCtxKeys sets the context keys to extract values from.
func SetCtxKeys(keys []string) {
	defaultLogger.SetCtxKeys(keys)
}
