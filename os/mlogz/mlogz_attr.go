package mlogz

type Attr struct {
	Key   string
	Value interface{}
}

// String creates a new string attribute.
func String(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

func Int(key string, value int) Attr {
	return Attr{Key: key, Value: value}
}

func Err(value error) Attr {
	return Attr{Key: "error", Value: value}
}
