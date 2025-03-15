package mlog

import (
	"reflect"

	"github.com/sirupsen/logrus"
)

// Hook defines the log hook interface.
type Hook interface {
	// Levels returns the log levels that the hook applies to.
	Levels() []Level
	// Fire executes the hook when a log entry is written.
	Fire(entry *Entry) error
}

// Entry represents a log entry.
type Entry struct {
	// log level
	Level Level
	// log message
	Message string
	// log fields
	Data map[string]interface{}
	// raw logrus entry (not exposed to users)
	raw *logrus.Entry
}

// logrusHook is an adapter for logrus.Hook.
type logrusHook struct {
	hook Hook
}

// Levels implements the logrus.Hook interface.
func (h *logrusHook) Levels() []logrus.Level {
	levels := h.hook.Levels()
	logrusLevels := make([]logrus.Level, len(levels))
	for i, level := range levels {
		logrusLevels[i] = logrus.Level(level)
	}
	return logrusLevels
}

// Fire implements the logrus.Hook interface.
func (h *logrusHook) Fire(entry *logrus.Entry) error {
	// create mlog.Entry
	e := &Entry{
		Level:   Level(entry.Level),
		Message: entry.Message,
		Data:    make(map[string]interface{}),
		raw:     entry,
	}

	// copy data
	for k, v := range entry.Data {
		e.Data[k] = v
	}

	// call user's hook
	return h.hook.Fire(e)
}

// AddHook adds a log hook.
func (l *Logger) AddHook(hook Hook) {
	l.parent.AddHook(&logrusHook{hook: hook})
}

// RemoveHookByType removes hooks of a specific type.
// It compares the type of the hook with the provided hook type.
func (l *Logger) RemoveHookByType(hookType Hook) {
	typeToRemove := reflect.TypeOf(hookType)

	// Iterate through all levels and hooks
	for level, hooks := range l.parent.Hooks {
		var newHooks []logrus.Hook

		// Keep only hooks that don't match the type to remove
		for _, hook := range hooks {
			if lh, ok := hook.(*logrusHook); ok {
				// Check if the underlying hook is of the type we want to remove
				if reflect.TypeOf(lh.hook) != typeToRemove {
					newHooks = append(newHooks, hook)
				}
			} else {
				// Keep hooks that are not logrusHook
				newHooks = append(newHooks, hook)
			}
		}

		// Update hooks for this level
		l.parent.Hooks[level] = newHooks
	}
}
