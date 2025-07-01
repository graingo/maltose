package mlog

import (
	"context"

	"github.com/graingo/maltose/errors/merror"
)

type Hook interface {
	// Name returns the name of the hook.
	Name() string
	// Level returns the level of the hook.
	Levels() []Level
	// Fire is called when the log is written.
	Fire(ctx context.Context, msg string, fields Fields) (string, Fields)
}

// AddHook adds a hook to the logger.
func (l *Logger) AddHook(hook Hook) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, h := range l.hooks {
		if h.Name() == hook.Name() {
			return merror.Newf("hook %s already exists", hook.Name())
		}
	}
	l.hooks = append(l.hooks, hook)
	return nil
}

// RemoveHook removes a hook from the logger.
func (l *Logger) RemoveHook(hookName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, h := range l.hooks {
		if h.Name() == hookName {
			l.hooks = append(l.hooks[:i], l.hooks[i+1:]...)
			break
		}
	}
}
