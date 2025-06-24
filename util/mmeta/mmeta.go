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
	metaAttributeName = "Meta"       // The attribute name of metadata in the structure
	metaTypeName      = "mmeta.Meta" // The type name for type string comparison
)

// Data retrieves and returns all metadata from `object`
func Data(object any) map[string]string {
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

// Get retrieves and returns the specified metadata from `object`
func Get(object any, key string) *mvar.Var {
	v, ok := Data(object)[key]
	if !ok {
		return nil
	}
	return mvar.New(v)
}

// StructType retrieves and returns the reflection type of the structure
func StructType(object any) (reflect.Type, error) {
	var reflectType reflect.Type
	if rt, ok := object.(reflect.Type); ok {
		reflectType = rt
	} else {
		reflectType = reflect.TypeOf(object)
	}
	if reflectType == nil {
		return nil, merror.New("invalid object type: nil")
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
	var (
		key  string
		data = make(map[string]string)
	)

	for tag != "" {
		// skip leading spaces
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// scan to colon
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		key = tag[:i]
		tag = tag[i+1:]

		// scan the value in quotes
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
