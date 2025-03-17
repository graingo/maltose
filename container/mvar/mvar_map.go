package mvar

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Map will convert the value to a map[string]any.
func (v *Var) Map() map[string]any {
	if v == nil {
		return nil
	}

	switch value := v.Val().(type) {
	case map[string]any:
		return value
	case map[any]any:
		m := make(map[string]any, len(value))
		for k, v := range value {
			m[String(k)] = v
		}
		return m
	case string:
		if value == "" {
			return nil
		}
		m := make(map[string]any)
		if json.Unmarshal([]byte(value), &m) == nil {
			return m
		}
	case []byte:
		if len(value) == 0 {
			return nil
		}
		m := make(map[string]any)
		if json.Unmarshal(value, &m) == nil {
			return m
		}
	case *map[string]any:
		if value == nil {
			return nil
		}
		return *value
	case *map[any]any:
		if value == nil {
			return nil
		}
		m := make(map[string]any, len(*value))
		for k, v := range *value {
			m[String(k)] = v
		}
		return m
	default:
		if value == nil {
			return nil
		}
		if v, ok := value.(map[string]any); ok {
			return v
		}
		rv := reflect.ValueOf(value)
		kind := rv.Kind()
		if kind == reflect.Ptr {
			rv = rv.Elem()
			kind = rv.Kind()
		}
		switch kind {
		case reflect.Map:
			m := make(map[string]any, rv.Len())
			for _, key := range rv.MapKeys() {
				m[String(key.Interface())] = rv.MapIndex(key).Interface()
			}
			return m
		case reflect.Struct:
			m := make(map[string]any)
			rt := rv.Type()
			for i := 0; i < rv.NumField(); i++ {
				field := rt.Field(i)
				if field.Anonymous {
					continue
				}
				m[field.Name] = rv.Field(i).Interface()
			}
			return m
		}
	}
	return nil
}

// String will convert any type to a string.
func String(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
