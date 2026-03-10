package schema

import (
	"encoding/json"
	"errors"
	"fmt"
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

func suggestType(t any) (string, error) {
	var suggestType string
	var err error

	switch val := t.(type) {
	case string:
		suggestType = "string"
	case bool:
		suggestType = "bool"
	case json.Number:
		if _, parseErr := val.Int64(); parseErr == nil {
			suggestType = "int"
		} else if _, parseErr := val.Float64(); parseErr == nil {
			suggestType = "float64"
		} else {
			err = errors.New("json.Number (unknown numeric)")
		}
	case map[string]any:
		suggestType = "struct"
	case []any:
		suggestType = "[]any"
	case nil:
		suggestType = "any (nil value)"
	default:
		err = fmt.Errorf("unknown type: %v", t)
	}

	return suggestType, err
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
