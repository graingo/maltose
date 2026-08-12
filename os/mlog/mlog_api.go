package mlog

import "context"

// Debugf prints the logging content with [DEBU] header, custom format and newline.
func Debugf(ctx context.Context, format string, v ...any) {
	DefaultLogger().Debugf(ctx, format, v...)
}

// Debugw prints the logging content with [DEBU] header, custom format and newline.
func Debugw(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger().Debugw(ctx, msg, fields...)
}

// Infof prints the logging content with [INFO] header, custom format and newline.
func Infof(ctx context.Context, format string, v ...any) {
	DefaultLogger().Infof(ctx, format, v...)
}

// Info prints the logging content with [INFO] header and newline.
func Infow(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger().Infow(ctx, msg, fields...)
}

// Warnf prints the logging content with [WARN] header, custom format and newline.
func Warnf(ctx context.Context, format string, v ...any) {
	DefaultLogger().Warnf(ctx, format, v...)
}

// Warnw prints the logging content with [WARN] header, custom format and newline.
func Warnw(ctx context.Context, msg string, fields ...Field) {
	DefaultLogger().Warnw(ctx, msg, fields...)
}

// Errorf prints the logging content with [ERRO] header, custom format and newline.
func Errorf(ctx context.Context, err error, format string, v ...any) {
	DefaultLogger().Errorf(ctx, err, format, v...)
}

// Errorw prints the logging content with [ERRO] header, custom format and newline.
func Errorw(ctx context.Context, err error, msg string, fields ...Field) {
	DefaultLogger().Errorw(ctx, err, msg, fields...)
}

// Fatalf prints the logging content with [FATA] header, custom format and newline, then exit the current process.
func Fatalf(ctx context.Context, err error, format string, v ...any) {
	DefaultLogger().Fatalf(ctx, err, format, v...)
}

// Fatalw prints the logging content with [FATA] header, custom format and newline, then exit the current process.
func Fatalw(ctx context.Context, err error, msg string, fields ...Field) {
	DefaultLogger().Fatalw(ctx, err, msg, fields...)
}

// Panicf prints the logging content with [PANI] header, custom format and newline, then panics.
func Panicf(ctx context.Context, err error, format string, v ...any) {
	DefaultLogger().Panicf(ctx, err, format, v...)
}

func Panicw(ctx context.Context, err error, msg string, fields ...Field) {
	DefaultLogger().Panicw(ctx, err, msg, fields...)
}

// With returns a new logger with the added attributes.
func With(fields ...Field) *Logger {
	return DefaultLogger().With(fields...)
}

// AddHook adds a hook to the logger.
func AddHook(hook Hook) {
	_ = DefaultLogger().AddHook(hook)
}

// RemoveHook removes a hook from the logger.
func RemoveHook(hookName string) {
	DefaultLogger().RemoveHook(hookName)
}

// Close closes the logger and its underlying resources.
func Close() error {
	return DefaultLogger().Close()
}
