package mvar

import (
	"encoding/json"
	"time"

	"github.com/spf13/cast"
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
	return cast.ToString(v.Val())
}

// Bool 将值转换为布尔型
func (v *Var) Bool() bool {
	if v == nil {
		return false
	}
	return cast.ToBool(v.Val())
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
	return cast.ToInt64(v.Val())
}

// Uint64 将值转换为 uint64
func (v *Var) Uint64() uint64 {
	if v == nil {
		return 0
	}
	return cast.ToUint64(v.Val())
}

// Float64 将值转换为 float64
func (v *Var) Float64() float64 {
	if v == nil {
		return 0
	}
	return cast.ToFloat64(v.Val())
}

// Time 将值转换为 time.Time
// format 参数用于指定时间字符串的格式
func (v *Var) Time(format ...string) time.Time {
	if v == nil {
		return time.Time{}
	}
	return cast.ToTime(v.Val())
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
