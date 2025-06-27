package mlogs

import (
	"log/slog"
	"strings"

	"github.com/graingo/maltose/errors/merror"
)

type Level slog.Level

const (
	DebugLevel = Level(slog.LevelDebug)
	InfoLevel  = Level(slog.LevelInfo)
	WarnLevel  = Level(slog.LevelWarn)
	ErrorLevel = Level(slog.LevelError)
	FatalLevel = Level(12)
	PanicLevel = Level(16)
)

// AllLevels returns all supported log levels.
func AllLevels() []Level {
	return []Level{DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, PanicLevel}
}

// ParseLevel parses a string level and returns the Level value.
func ParseLevel(level string) (Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "panic":
		return PanicLevel, nil
	default:
		return 0, merror.Newf("invalid log level: %s", level)
	}
}

func (l Level) toSlogLevel() slog.Level {
	return slog.Level(l)
}
