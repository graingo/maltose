package mvar

import "reflect"

// IsNil checks if the value is nil.
func (v *Var) IsNil() bool {
	return v == nil || v.value == nil
}

// IsEmpty checks if the value is empty.
func (v *Var) IsEmpty() bool {
	if v == nil {
		return true
	}
	switch value := v.Val().(type) {
	case int:
		return value == 0
	case int8:
		return value == 0
	case int16:
		return value == 0
	case int32:
		return value == 0
	case int64:
		return value == 0
	case uint:
		return value == 0
	case uint8:
		return value == 0
	case uint16:
		return value == 0
	case uint32:
		return value == 0
	case uint64:
		return value == 0
	case float32:
		return value == 0
	case float64:
		return value == 0
	case bool:
		return !value
	case string:
		return value == ""
	case []byte:
		return len(value) == 0
	case []rune:
		return len(value) == 0
	case []int:
		return len(value) == 0
	case []string:
		return len(value) == 0
	case map[string]any:
		return len(value) == 0
	case map[any]any:
		return len(value) == 0
	default:
		return v.IsNil()
	}
}

// IsInt checks if the value can be converted to an integer.
func (v *Var) IsInt() bool {
	switch v.Val().(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

// IsUint checks if the value can be converted to an unsigned integer.
func (v *Var) IsUint() bool {
	switch v.Val().(type) {
	case uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

// IsFloat checks if the value can be converted to a float.
func (v *Var) IsFloat() bool {
	switch v.Val().(type) {
	case float32, float64:
		return true
	}
	return false
}

// IsSlice checks if the value is a slice type.
func (v *Var) IsSlice() bool {
	switch v.Val().(type) {
	case []interface{}, []int, []string, []byte, []rune:
		return true
	}
	return false
}

// IsMap checks if the value is a map type.
func (v *Var) IsMap() bool {
	switch v.Val().(type) {
	case map[string]interface{}, map[interface{}]interface{}:
		return true
	}
	return false
}

// IsStruct checks if the value is a struct type.
func (v *Var) IsStruct() bool {
	if v.IsNil() {
		return false
	}
	return reflect.TypeOf(v.Val()).Kind() == reflect.Struct
}
