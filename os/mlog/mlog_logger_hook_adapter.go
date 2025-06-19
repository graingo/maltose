package mlog

import (
	"maps"

	"github.com/sirupsen/logrus"
)

// logrusHook is an adapter that wraps a mlog.Hook to make it compatible with the logrus.Hook interface.
// This allows mlog to use its own simplified Hook interface while still leveraging the underlying logrus hooking mechanism.
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
		Data:    make(map[string]any),
		raw:     entry,
	}

	// copy data
	maps.Copy(e.Data, entry.Data)

	// call user's hook
	return h.hook.Fire(e)
}
