package mmeta

import (
	"reflect"
	"strconv"

	"github.com/graingo/maltose/container/mvar"
	"github.com/graingo/maltose/errors/merror"
)

// Meta is used as an embedded property to enable metadata features
type Meta struct{}

const (
	metaAttributeName = "Meta" // The attribute name of metadata in the structure
)

// metaType holds the reflection type of Meta, used for efficient type comparison.
var metaType = reflect.TypeOf(Meta{})

// Data returns all metadata of `object` as map.
// The parameter `object` can be any struct object or pointer to struct.
func Data(object any) map[string]string {
	if object == nil {
		return map[string]string{}
	}
	
	reflectType, err := StructType(object)
	if err != nil {
		return map[string]string{}
	}
	
	if field, ok := reflectType.FieldByName(metaAttributeName); ok {
		if field.Type == metaType {
			return ParseTag(string(field.Tag))
		}
	}
	return map[string]string{}
}

// Get returns the value of the specified metadata by key.
// The parameter `object` can be any struct object or pointer to struct.
func Get(object any, key string) *mvar.Var {
	data := Data(object)
	if v, ok := data[key]; ok {
		return mvar.New(v)
	}
	return nil
}

// StructType retrieves and returns the reflection type of the structure
func StructType(object any) (reflect.Type, error) {
	if object == nil {
		return nil, merror.New("invalid object type: nil")
	}
	
	var reflectType reflect.Type
	if rt, ok := object.(reflect.Type); ok {
		reflectType = rt
	} else {
		v := reflect.ValueOf(object)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return nil, merror.New("invalid object: nil pointer")
		}
		reflectType = v.Type()
	}
	
	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType.Kind() != reflect.Struct {
		return nil, merror.Newf("invalid object kind: %v", reflectType.Kind())
	}
	return reflectType, nil
}

// ParseTag parses the tag string to a map
func ParseTag(tag string) map[string]string {
	data := make(map[string]string)
	
	for tag != "" {
		// Skip leading spaces
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		key := tag[:i]
		tag = tag[i+1:]

		// Scan the value in quotes
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
