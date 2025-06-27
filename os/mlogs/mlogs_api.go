package mlogs

import "context"

// SetConfig applies configuration to the default logger.
func SetConfig(config *Config) error {
	return defaultLogger.SetConfig(config)
}

// SetConfigWithMap applies configuration from a map to the default logger.
func SetConfigWithMap(configMap map[string]any) error {
	config := defaultConfig()
	if err := config.SetConfigWithMap(configMap); err != nil {
		return err
	}
	return defaultLogger.SetConfig(config)
}

// --- Contextual Logger Creation ---

func With(attrs ...Attr) *Logger {
	return defaultLogger.With(attrs...)
}

// --- Global Logging Functions ---

func Debug(ctx context.Context, msg string, attrs ...Attr) {
	defaultLogger.Debug(ctx, msg, attrs...)
}

func Info(ctx context.Context, msg string, attrs ...Attr) {
	defaultLogger.Info(ctx, msg, attrs...)
}

func Warn(ctx context.Context, msg string, attrs ...Attr) {
	defaultLogger.Warn(ctx, msg, attrs...)
}

func Error(ctx context.Context, err error, msg string, attrs ...Attr) {
	defaultLogger.Error(ctx, err, msg, attrs...)
}

func Fatal(ctx context.Context, err error, msg string, attrs ...Attr) {
	defaultLogger.Fatal(ctx, err, msg, attrs...)
}

func Panic(ctx context.Context, err error, msg string, attrs ...Attr) {
	defaultLogger.Panic(ctx, err, msg, attrs...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	defaultLogger.Debugf(ctx, format, args...)
}

func Infof(ctx context.Context, format string, args ...any) {
	defaultLogger.Infof(ctx, format, args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	defaultLogger.Warnf(ctx, format, args...)
}

func Errorf(ctx context.Context, err error, format string, args ...any) {
	defaultLogger.Errorf(ctx, err, format, args...)
}

func Fatalf(ctx context.Context, err error, format string, args ...any) {
	defaultLogger.Fatalf(ctx, err, format, args...)
}

func Panicf(ctx context.Context, err error, format string, args ...any) {
	defaultLogger.Panicf(ctx, err, format, args...)
}
