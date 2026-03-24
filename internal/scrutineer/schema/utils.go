package schema

import (
	"encoding/json"
	"reflect"
)

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

func suggestJSONType(t any) DataType {
	var suggestType DataType

	switch val := t.(type) {
	case string:
		suggestType = String
	case bool:
		suggestType = Bool
	case json.Number:
		if _, parseErr := val.Int64(); parseErr == nil {
			suggestType = Int
		} else if _, parseErr := val.Float64(); parseErr == nil {
			suggestType = Float64
		} else {
			suggestType = UnknownNumeric
		}
	case map[string]any:
		suggestType = Object
	case []any:
		suggestType = Slice
	case nil:
		suggestType = Any
	default:
		suggestType = Unknown
	}

	return suggestType
}

func suggestGoType(val reflect.Value) DataType {
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return Any
		}
		val = val.Elem()
	}

	var suggestType DataType

	switch val.Kind() {
	case reflect.String:
		suggestType = String
	case reflect.Bool:
		suggestType = Bool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		suggestType = Int
	case reflect.Float32, reflect.Float64:
		suggestType = Float64
	case reflect.Map, reflect.Struct:
		suggestType = Object
	case reflect.Slice, reflect.Array:
		suggestType = Slice
	case reflect.Interface:
		suggestType = Any
	default:
		suggestType = Unknown
	}

	return suggestType
}

func normalizeJSONNumber(val any) (any, error) {
	// figure out if it is a json.Number because they appear as a string type,
	// which is bad for checking the zero value later since a 0 would appear as "0"
	if num, isNum := val.(json.Number); isNum {
		// we parse it as a float64 because every number can safely be parsed as float64
		// this will have no effect on type checking since here we are only interested in the value
		return num.Float64()
	}
	return val, nil
}
