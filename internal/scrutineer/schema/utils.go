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
