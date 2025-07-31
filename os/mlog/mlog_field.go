package mlog

import (
	"math"
	"time"

	"go.uber.org/zap/zapcore"
)

// Field is a key-value pair used for structured logging. It is a type-safe
// and optimized representation of a logging field, designed to be structurally
// compatible with a popular high-performance logging library's internal field type
// for zero-overhead conversion.
type Field struct {
	Key       string
	Type      zapcore.FieldType
	Integer   int64
	String    string
	Interface any
}

// Fields is a slice of Field.
type Fields []Field

// Any takes a key and an arbitrary value and chooses the best way to represent
// them as a field.
func Any(key string, value any) Field {
	// In a real high-performance scenario, you might add type switches here
	// for common types to avoid reflection, similar to what zap itself does.
	// For this encapsulation, relying on zapcore.ReflectType is a good balance.
	return Field{Key: key, Type: zapcore.ReflectType, Interface: value}
}

// Bool constructs a field with a boolean value.
func Bool(key string, val bool) Field {
	var intVal int64
	if val {
		intVal = 1
	}
	return Field{Key: key, Type: zapcore.BoolType, Integer: intVal}
}

// Err constructs a field with an error. If the error is nil, a no-op
// field is returned, which is ignored by the logger.
func Err(err error) Field {
	if err == nil {
		return Skip()
	}
	return Field{Key: "error", Type: zapcore.ErrorType, Interface: err}
}

// Duration constructs a field with a time.Duration.
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Type: zapcore.DurationType, Integer: int64(val)}
}

// Float64 constructs a field with a float64 value.
func Float64(key string, val float64) Field {
	return Field{Key: key, Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val))}
}

// Int constructs a field with an int value.
func Int(key string, val int) Field {
	return Field{Key: key, Type: zapcore.Int64Type, Integer: int64(val)}
}

// Int64 constructs a field with an int64 value.
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: zapcore.Int64Type, Integer: val}
}

// String constructs a field with a string value.
func String(key string, val string) Field {
	return Field{Key: key, Type: zapcore.StringType, String: val}
}

// Time constructs a field with a time.Time value. It's formatted as a
// floating-point number of seconds since the Unix epoch.
func Time(key string, val time.Time) Field {
	return Field{Key: key, Type: zapcore.TimeType, Integer: val.UnixNano(), Interface: time.UTC}
}

// Uint constructs a field with an unsigned integer value.
func Uint(key string, val uint) Field {
	return Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(val)}
}

// Uint64 constructs a field with a uint64 value.
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: zapcore.Uint64Type, Integer: int64(val)}
}

// Skip constructs a no-op field, which is ignored by the logger.
func Skip() Field {
	return Field{Type: zapcore.SkipType}
}
