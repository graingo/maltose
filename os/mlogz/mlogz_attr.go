package mlogz

import "time"

type Attr struct {
	Key   string
	Value any
}

// String creates a new string attribute.
func String(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

// Int creates a new int attribute.
func Int(key string, value int) Attr {
	return Attr{Key: key, Value: value}
}

// Uint creates a new uint attribute.
func Uint(key string, value uint) Attr {
	return Attr{Key: key, Value: value}
}

// Int64 creates a new int64 attribute.
func Int64(key string, value int64) Attr {
	return Attr{Key: key, Value: value}
}

// Uint64 creates a new uint64 attribute.
func Uint64(key string, value uint64) Attr {
	return Attr{Key: key, Value: value}
}

// Float64 creates a new float64 attribute.
func Float64(key string, value float64) Attr {
	return Attr{Key: key, Value: value}
}

// Bool creates a new bool attribute.
func Bool(key string, value bool) Attr {
	return Attr{Key: key, Value: value}
}

// Time creates a new time attribute.
func Time(key string, value time.Time) Attr {
	return Attr{Key: key, Value: value}
}

// Err creates a new error attribute.
func Err(value error) Attr {
	return Attr{Key: "error", Value: value}
}

// Any creates a new any attribute.
func Any(key string, value any) Attr {
	return Attr{Key: key, Value: value}
}
