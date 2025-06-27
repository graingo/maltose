package mvar

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/graingo/mconv"
)

// Var is a universal variable type implementation.
type Var struct {
	value any  // The underlying value.
	safe  bool // Whether to enable thread safety, default is false.
	mu    sync.RWMutex
}

// New creates and returns a new Var.
// The safe parameter is used to specify whether to enable thread safety, the default is false.
func New(value any, safe ...bool) *Var {
	if len(safe) > 0 && safe[0] {
		return &Var{
			value: value,
			safe:  true,
			mu:    sync.RWMutex{},
		}
	}
	return &Var{
		value: value,
	}
}

// Val returns the original value.
func (v *Var) Val() any {
	if v == nil {
		return nil
	}
	return v.value
}

// Interface is an alias for Val.
func (v *Var) Interface() any {
	return v.Val()
}

// String converts the value to a string.
func (v *Var) String() string {
	if v == nil {
		return ""
	}
	return mconv.ToString(v.Val())
}

// Bool converts the value to a boolean.
func (v *Var) Bool() bool {
	if v == nil {
		return false
	}
	return mconv.ToBool(v.Val())
}

// Int converts the value to an int.
func (v *Var) Int() int {
	return int(v.Int64())
}

// Int64 converts the value to an int64.
func (v *Var) Int64() int64 {
	if v == nil {
		return 0
	}
	return mconv.ToInt64(v.Val())
}

// Uint64 converts the value to an uint64.
func (v *Var) Uint64() uint64 {
	if v == nil {
		return 0
	}
	return mconv.ToUint64(v.Val())
}

// Float64 converts the value to a float64.
func (v *Var) Float64() float64 {
	if v == nil {
		return 0
	}
	return mconv.ToFloat64(v.Val())
}

// Time converts the value to a time.Time.
// The format parameter is used to specify the format of the time string.
func (v *Var) Time(format ...string) time.Time {
	if v == nil {
		return time.Time{}
	}
	return mconv.ToTime(v.Val())
}

// Struct maps the value to a struct.
// The `pointer` parameter should be a pointer to a struct.
// The `hooks` parameter is used to specify the hooks for the conversion.
func (v *Var) Struct(pointer any, hooks ...mconv.HookFunc) error {
	if v == nil {
		return nil
	}
	return mconv.ToStructE(v.Val(), pointer, hooks...)
}

// MarshalJSON implements the json.Marshaler interface.
func (v *Var) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (v *Var) UnmarshalJSON(b []byte) error {
	var i any
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}
	v.value = i
	return nil
}
