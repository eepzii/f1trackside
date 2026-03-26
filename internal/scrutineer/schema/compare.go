package schema

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

func (f *Field) Compare(minimumType any, responseTypeVal any) {
	v := reflect.ValueOf(responseTypeVal)

	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v = reflect.New(v.Type().Elem()).Elem()
			continue
		}
		v = v.Elem()
	}

	var minimumTypeMap map[string]any

	switch data := minimumType.(type) {
	case map[string]any:
		minimumTypeMap = data
	case []any:
		minimumTypeMap = make(map[string]any)
		for i, val := range data {
			minimumTypeMap[fmt.Sprintf("%d", i)] = val
		}
	default:
		jsonType := suggestJSONType(data)
		goType := suggestGoType(v)

		if goType != Any && goType != Unknown && goType != jsonType {
			msg := fmt.Sprintf("want: %s, got: %s", jsonType, v.Type().String())
			if f.Errors == nil {
				f.Errors = make(map[string]int)
			}
			f.Errors[msg]++
		}
		return
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		t := v.Type().Elem()
		template := reflect.New(t).Elem().Interface()

		for _, entry := range minimumTypeMap {
			f.Compare(entry, template)
		}
	case reflect.Struct:
		var usedKeys []string

		t := v.Type()
		for i := range t.NumField() {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			fieldVal := v.Field(i)

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			key := strings.Split(jsonTag, ",")[0]
			usedKeys = append(usedKeys, key)

			if f.Children[key] == nil {
				f.Children[key] = &Field{
					Name:     key,
					Children: make(map[string]*Field),
					Errors:   make(map[string]int),
				}
			}
			childStat := f.Children[key]

			entry, exists := minimumTypeMap[key]
			if !exists {
				continue
			}
			childStat.total++

			jsonType := suggestJSONType(entry)
			if !slices.Contains(childStat.observedTypes, jsonType) {
				childStat.observedTypes = append(childStat.observedTypes, jsonType)
			}

			goType := suggestGoType(fieldVal)

			isDynamicArray := jsonType == Slice && goType == Object
			if goType != Any && goType != Unknown && goType != jsonType && !isDynamicArray {
				msg := fmt.Sprintf("want: %s, got: %s", jsonType, fieldVal.Type().String())
				childStat.Errors[msg]++
			}

			childStat.checkUnsafeDefaults(field.Type, fieldVal, entry)

			childStat.Compare(entry, fieldVal.Interface())
		}

		f.checkMissingFields(usedKeys, minimumTypeMap)
	default:
		return
	}
}
