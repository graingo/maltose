package mlog

import "context"

type Entry struct {
	ctx    context.Context
	msg    string
	fields Fields
}

func (e *Entry) AddField(field Field) *Entry {
	e.fields = append(e.fields, field)
	return e
}

func (e *Entry) SetMsg(msg string) *Entry {
	e.msg = msg
	return e
}

func (e *Entry) GetMsg() string {
	return e.msg
}

func (e *Entry) GetFields() Fields {
	return e.fields
}

func (e *Entry) GetContext() context.Context {
	return e.ctx
}
