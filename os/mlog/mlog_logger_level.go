package mlog

import (
	"github.com/sirupsen/logrus"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
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
