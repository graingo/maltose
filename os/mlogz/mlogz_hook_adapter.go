package mlogz

import "context"

type Hook interface {
	// Level returns the level of the hook.
	Level() Level
	// Fire is called when the log is written.
	Fire(ctx context.Context, msg string, attrs []Attr) (string, []Attr)
}
