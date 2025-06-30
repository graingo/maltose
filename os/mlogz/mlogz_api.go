package mlogz

import "context"

// Debug prints the logging content with [DEBU] header and newline.
func Debug(ctx context.Context, msg string) {
	defaultLogger.Debug(ctx, msg)
}

// Debugf prints the logging content with [DEBU] header, custom format and newline.
func Debugf(ctx context.Context, format string, v ...any) {
	defaultLogger.Debugf(ctx, format, v...)
}

// Debugw prints the logging content with [DEBU] header, custom format and newline.
func Debugw(ctx context.Context, msg string, attrs ...Attr) {
	defaultLogger.Debugw(ctx, msg, attrs...)
}

// Info prints the logging content with [INFO] header and newline.
func Info(ctx context.Context, msg string) {
	defaultLogger.Info(ctx, msg)
}

// Infof prints the logging content with [INFO] header, custom format and newline.
func Infof(ctx context.Context, format string, v ...any) {
	defaultLogger.Infof(ctx, format, v...)
}

// Info prints the logging content with [INFO] header and newline.
func Infow(ctx context.Context, msg string, attrs ...Attr) {
	defaultLogger.Infow(ctx, msg, attrs...)
}

// Warn prints the logging content with [WARN] header and newline.
func Warn(ctx context.Context, msg string) {
	defaultLogger.Warn(ctx, msg)
}

// Warnf prints the logging content with [WARN] header, custom format and newline.
func Warnf(ctx context.Context, format string, v ...any) {
	defaultLogger.Warnf(ctx, format, v...)
}

// Warnw prints the logging content with [WARN] header, custom format and newline.
func Warnw(ctx context.Context, msg string, attrs ...Attr) {
	defaultLogger.Warnw(ctx, msg, attrs...)
}

// Error prints the logging content with [ERRO] header and newline.
func Error(ctx context.Context, err error, msg string) {
	defaultLogger.Error(ctx, err, msg)
}

// Errorf prints the logging content with [ERRO] header, custom format and newline.
func Errorf(ctx context.Context, err error, format string, v ...any) {
	defaultLogger.Errorf(ctx, err, format, v...)
}

// Errorw prints the logging content with [ERRO] header, custom format and newline.
func Errorw(ctx context.Context, err error, msg string, attrs ...Attr) {
	defaultLogger.Errorw(ctx, err, msg, attrs...)
}

// Fatal prints the logging content with [FATA] header and newline, then exit the current process.
func Fatal(ctx context.Context, err error, msg string) {
	defaultLogger.Fatal(ctx, err, msg)
}

// Fatalf prints the logging content with [FATA] header, custom format and newline, then exit the current process.
func Fatalf(ctx context.Context, err error, format string, v ...any) {
	defaultLogger.Fatalf(ctx, err, format, v...)
}

// Fatalw prints the logging content with [FATA] header, custom format and newline, then exit the current process.
func Fatalw(ctx context.Context, err error, msg string, attrs ...Attr) {
	defaultLogger.Fatalw(ctx, err, msg, attrs...)
}

// Panic prints the logging content with [PANI] header and newline, then panics.
func Panic(ctx context.Context, err error, msg string) {
	defaultLogger.Panic(ctx, err, msg)
}

// Panicf prints the logging content with [PANI] header, custom format and newline, then panics.
func Panicf(ctx context.Context, err error, format string, v ...any) {
	defaultLogger.Panicf(ctx, err, format, v...)
}

func Panicw(ctx context.Context, err error, msg string, attrs ...Attr) {
	defaultLogger.Panicw(ctx, err, msg, attrs...)
}

func With(attrs ...Attr) *Logger {
	return defaultLogger.With(attrs...)
}

func AddHook(hook Hook) {
	defaultLogger.AddHook(hook)
}
