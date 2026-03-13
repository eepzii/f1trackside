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

func suggestType(t any) DataType {
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
