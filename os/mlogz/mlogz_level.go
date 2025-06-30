package mlogz

import (
	"strings"

	"github.com/graingo/maltose/errors/merror"
	"go.uber.org/zap/zapcore"
)

// Level is the log level.
type Level int8

const (
	DebugLevel Level = Level(zapcore.DebugLevel)
	InfoLevel  Level = Level(zapcore.InfoLevel)
	WarnLevel  Level = Level(zapcore.WarnLevel)
	ErrorLevel Level = Level(zapcore.ErrorLevel)
	FatalLevel Level = Level(zapcore.FatalLevel)
	PanicLevel Level = Level(zapcore.PanicLevel)
)

func AllLevels() []Level {
	return []Level{
		DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
		PanicLevel,
	}
}

// SetLevel sets the logging level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the logging level value.
func (l *Logger) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
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

// toZapLevel converts Level to zapcore.Level.
func (l Level) toZapLevel() zapcore.Level {
	return zapcore.Level(l)
}
