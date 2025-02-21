// Package gvar 提供通用变量类型
package gvar

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Var 是一个通用变量类型的实现
type Var struct {
	value interface{} // 底层值
	safe  bool        // 是否线程安全
}

// New 创建并返回一个新的 Var
// safe 参数用于指定是否启用线程安全，默认为 false
func New(value interface{}, safe ...bool) *Var {
	if len(safe) > 0 && safe[0] {
		return &Var{
			value: value,
			safe:  true,
		}
	}
	return &Var{
		value: value,
	}
}

// Val 返回原始值
func (v *Var) Val() interface{} {
	if v == nil {
		return nil
	}
	return v.value
}

// Interface 是 Val 的别名
func (v *Var) Interface() interface{} {
	return v.Val()
}

// String 将值转换为字符串
func (v *Var) String() string {
	if v == nil {
		return ""
	}
	switch result := v.Val().(type) {
	case string:
		return result
	case []byte:
		return string(result)
	case int:
		return strconv.Itoa(result)
	case int64:
		return strconv.FormatInt(result, 10)
	case float64:
		return strconv.FormatFloat(result, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(result)
	default:
		if result == nil {
			return ""
		}
		return fmt.Sprintf("%v", result)
	}
}

// Bool 将值转换为布尔型
func (v *Var) Bool() bool {
	if v == nil {
		return false
	}
	switch value := v.Val().(type) {
	case bool:
		return value
	case string:
		str := value
		if str != "" && str != "0" && str != "false" {
			return true
		}
		return false
	case int, int8, int16, int32, int64:
		return v.Int64() != 0
	case uint, uint8, uint16, uint32, uint64:
		return v.Uint64() != 0
	case float32, float64:
		return v.Float64() != 0
	default:
		return false
	}
}

// Int 将值转换为 int
func (v *Var) Int() int {
	return int(v.Int64())
}

// Int64 将值转换为 int64
func (v *Var) Int64() int64 {
	if v == nil {
		return 0
	}
	switch value := v.Val().(type) {
	case int:
		return int64(value)
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case uint:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case uint64:
		return int64(value)
	case float32:
		return int64(value)
	case float64:
		return int64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	case string:
		val, _ := strconv.ParseInt(value, 10, 64)
		return val
	default:
		return 0
	}
}

// Uint64 将值转换为 uint64
func (v *Var) Uint64() uint64 {
	if v == nil {
		return 0
	}
	switch value := v.Val().(type) {
	case int:
		return uint64(value)
	case int8:
		return uint64(value)
	case int16:
		return uint64(value)
	case int32:
		return uint64(value)
	case int64:
		return uint64(value)
	case uint:
		return uint64(value)
	case uint8:
		return uint64(value)
	case uint16:
		return uint64(value)
	case uint32:
		return uint64(value)
	case uint64:
		return value
	case float32:
		return uint64(value)
	case float64:
		return uint64(value)
	case bool:
		if value {
			return 1
		}
		return 0
	case string:
		val, _ := strconv.ParseUint(value, 10, 64)
		return val
	default:
		return 0
	}
}

// Float64 将值转换为 float64
func (v *Var) Float64() float64 {
	if v == nil {
		return 0
	}
	switch value := v.Val().(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case string:
		val, _ := strconv.ParseFloat(value, 64)
		return val
	default:
		return 0
	}
}

// Time 将值转换为 time.Time
// format 参数用于指定时间字符串的格式
func (v *Var) Time(format ...string) time.Time {
	if v == nil {
		return time.Time{}
	}
	switch value := v.Val().(type) {
	case time.Time:
		return value
	case string:
		if len(format) > 0 {
			t, _ := time.Parse(format[0], value)
			return t
		}
		t, _ := time.Parse(time.RFC3339, value)
		return t
	case int64:
		return time.Unix(value, 0)
	default:
		return time.Time{}
	}
}

// MarshalJSON 实现 json.Marshaler 接口
func (v *Var) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Val())
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (v *Var) UnmarshalJSON(b []byte) error {
	var i interface{}
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}
	v.value = i
	return nil
}
