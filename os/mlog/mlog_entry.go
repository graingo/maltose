package mlog

import (
	"context"
	"sync"
)

// entryPool is a pool of Entry instances.
var entryPool = sync.Pool{
	New: func() any {
		return &Entry{}
	},
}

// Entry represents a log entry.
type Entry struct {
	ctx    context.Context
	msg    string
	fields Fields
}

// SetMsg sets the message of the entry.
func (e *Entry) SetMsg(msg string) *Entry {
	e.msg = msg
	return e
}

// GetMsg returns the message of the entry.
func (e *Entry) GetMsg() string {
	return e.msg
}

// AddField adds a field to the entry.
func (e *Entry) AddField(field Field) *Entry {
	e.fields = append(e.fields, field)
	return e
}

// GetFields returns the fields of the entry.
func (e *Entry) GetFields() Fields {
	return e.fields
}

// SetFields sets the fields of the entry.
func (e *Entry) SetFields(fields Fields) *Entry {
	e.fields = fields
	return e
}

// GetContext returns the context of the entry.
func (e *Entry) GetContext() context.Context {
	return e.ctx
}

// SetContext sets the context of the entry.
func (e *Entry) SetContext(ctx context.Context) *Entry {
	e.ctx = ctx
	return e
}

// Reset resets the entry.
func (e *Entry) reset() *Entry {
	e.ctx = nil
	e.msg = ""
	e.fields = nil
	return e
}
