package mlog

import "context"

// Print prints `v` with newline using fmt.Sprintln.
// The parameter `v` can be multiple variables.
func Print(ctx context.Context, v ...interface{}) {
	defaultLogger.Print(ctx, v...)
}

// Printf prints `v` with format `format` using fmt.Sprintf.
// The parameter `v` can be multiple variables.
func Printf(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Printf(ctx, format, v...)
}

// Fatal prints the logging content with [FATA] header and newline, then exit the current process.
func Fatal(ctx context.Context, v ...interface{}) {
	defaultLogger.Fatal(ctx, v...)
}

// Fatalf prints the logging content with [FATA] header, custom format and newline, then exit the current process.
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Fatalf(ctx, format, v...)
}

// Panic prints the logging content with [PANI] header and newline, then panics.
func Panic(ctx context.Context, v ...interface{}) {
	defaultLogger.Panic(ctx, v...)
}

// Panicf prints the logging content with [PANI] header, custom format and newline, then panics.
func Panicf(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Panicf(ctx, format, v...)
}

// Info prints the logging content with [INFO] header and newline.
func Info(ctx context.Context, v ...interface{}) {
	defaultLogger.Info(ctx, v...)
}

// Infof prints the logging content with [INFO] header, custom format and newline.
func Infof(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Infof(ctx, format, v...)
}

// Debug prints the logging content with [DEBU] header and newline.
func Debug(ctx context.Context, v ...interface{}) {
	defaultLogger.Debug(ctx, v...)
}

// Debugf prints the logging content with [DEBU] header, custom format and newline.
func Debugf(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Debugf(ctx, format, v...)
}

// Warn prints the logging content with [WARN] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Warn(ctx context.Context, v ...interface{}) {
	defaultLogger.Warn(ctx, v...)
}

// Warnf prints the logging content with [WARN] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Warnf(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Warnf(ctx, format, v...)
}

// Error prints the logging content with [ERRO] header and newline.
// It also prints caller stack info if stack feature is enabled.
func Error(ctx context.Context, v ...interface{}) {
	defaultLogger.Error(ctx, v...)
}

// Errorf prints the logging content with [ERRO] header, custom format and newline.
// It also prints caller stack info if stack feature is enabled.
func Errorf(ctx context.Context, format string, v ...interface{}) {
	defaultLogger.Errorf(ctx, format, v...)
}
