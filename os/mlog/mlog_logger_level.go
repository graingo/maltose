package mlog

import (
	"strings"

	"github.com/graingo/maltose/errors/merror"
	"github.com/sirupsen/logrus"
)

type Level int

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
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
	l.parent.SetLevel(logrus.Level(level))
}

// GetLevel returns the logging level value.
func (l *Logger) GetLevel() Level {
	return Level(l.parent.GetLevel())
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
