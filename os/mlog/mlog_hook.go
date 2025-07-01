package mlog

import (
	"github.com/graingo/maltose/errors/merror"
)

type Hook interface {
	// Name returns the name of the hook.
	Name() string
	// Level returns the level of the hook.
	Levels() []Level
	// Fire is called when the log is written.
	Fire(entry *Entry)
}

// AddHook adds a hook to the logger.
func (l *Logger) AddHook(hook Hook) error {
	l.hookMu.Lock()
	defer l.hookMu.Unlock()
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
	l.hookMu.Lock()
	defer l.hookMu.Unlock()
	for i, h := range l.hooks {
		if h.Name() == hookName {
			l.hooks = append(l.hooks[:i], l.hooks[i+1:]...)
			break
		}
	}

}
