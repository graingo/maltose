// Package qmeta provides embedded meta data feature for struct.
package qmeta

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Meta 用作结构体的嵌入属性以启用元数据特性
type Meta struct{}

// Var 是一个通用变量接口
type Var struct {
	value interface{}
}

const (
	metaAttributeName = "Meta"       // 结构体中元数据的属性名
	metaTypeName      = "qmeta.Meta" // 用于类型字符串比较
)

// Data 从 `object` 中检索并返回所有元数据
func Data(object interface{}) map[string]string {
	reflectType, err := StructType(object)
	if err != nil {
		return nil
	}
	if field, ok := reflectType.FieldByName(metaAttributeName); ok {
		if field.Type.String() == metaTypeName {
			return ParseTag(string(field.Tag))
		}
	}
	return map[string]string{}
}

// Get 从 `object` 中检索并返回指定的元数据
func Get(object interface{}, key string) *Var {
	v, ok := Data(object)[key]
	if !ok {
		return nil
	}
	return New(v)
}

// New 创建一个新的 Var
func New(value interface{}) *Var {
	return &Var{value: value}
}

// String 将值转换为字符串
func (v *Var) String() string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v.value))
}

// IsNil 检查值是否为 nil
func (v *Var) IsNil() bool {
	return v == nil || v.value == nil
}

// IsEmpty 检查值是否为空
func (v *Var) IsEmpty() bool {
	if v == nil {
		return true
	}
	return v.value == nil || v.String() == ""
}

// StructType 获取结构体的反射类型
func StructType(object interface{}) (reflect.Type, error) {
	var reflectType reflect.Type
	if rt, ok := object.(reflect.Type); ok {
		reflectType = rt
	} else {
		reflectType = reflect.TypeOf(object)
	}
	if reflectType == nil {
		return nil, fmt.Errorf("invalid object type: nil")
	}
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid object kind: %v", reflectType.Kind())
	}
	return reflectType, nil
}

// ParseTag 解析标签字符串为 map
func ParseTag(tag string) map[string]string {
	var (
		key  string
		data = make(map[string]string)
	)

	for tag != "" {
		// 跳过前导空格
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// 扫描到冒号
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		key = tag[:i]
		tag = tag[i+1:]

		// 扫描引号内的值
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		quotedValue := tag[:i+1]
		tag = tag[i+1:]
		value, err := strconv.Unquote(quotedValue)
		if err != nil {
			break
		}
		data[key] = value
	}
	return data
}
