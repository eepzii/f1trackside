package schema

import (
	"encoding/json"
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
		if _, err := val.Int64(); err == nil {
			suggestType = "int"
		} else if _, err := val.Float64(); err == nil {
			suggestType = "float64"
		} else {
			err = fmt.Errorf("json.Number (unknown numeric)")
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
