package mvar

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// Map 将值转换为 map[string]interface{}
func (v *Var) Map() map[string]interface{} {
	if v == nil {
		return nil
	}

	switch value := v.Val().(type) {
	case map[string]interface{}:
		return value
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(value))
		for k, v := range value {
			m[String(k)] = v
		}
		return m
	case string:
		// 尝试解析 JSON 字符串
		if value == "" {
			return nil
		}
		m := make(map[string]interface{})
		if json.Unmarshal([]byte(value), &m) == nil {
			return m
		}
	case []byte:
		if len(value) == 0 {
			return nil
		}
		m := make(map[string]interface{})
		if json.Unmarshal(value, &m) == nil {
			return m
		}
	case *map[string]interface{}:
		if value == nil {
			return nil
		}
		return *value
	case *map[interface{}]interface{}:
		if value == nil {
			return nil
		}
		m := make(map[string]interface{}, len(*value))
		for k, v := range *value {
			m[String(k)] = v
		}
		return m
	default:
		// 处理结构体
		if value == nil {
			return nil
		}
		if v, ok := value.(map[string]interface{}); ok {
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
			m := make(map[string]interface{}, rv.Len())
			for _, key := range rv.MapKeys() {
				m[String(key.Interface())] = rv.MapIndex(key).Interface()
			}
			return m
		case reflect.Struct:
			m := make(map[string]interface{})
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

// String 将任意类型转换为字符串
func String(value interface{}) string {
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
