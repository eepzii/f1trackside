package schema

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

func (f *Field) Compare(minimumType any, responseTypeVal any) {
	v := reflect.ValueOf(responseTypeVal)

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
		return
	}

	switch v.Kind() {
	case reflect.Map:
		iter := v.MapRange()

		for iter.Next() {
			key := iter.Key().String()
			val := iter.Value()

			if entry, exists := minimumTypeMap[key]; exists {
				f.Compare(entry, val.Interface())
			}
		}
	case reflect.Struct:
		var usedKeys []string

		t := v.Type()
		for i := range t.NumField() {
			field := t.Field(i)
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
				}
			}
			childStat := f.Children[key]

			entry, exists := minimumTypeMap[key]
			if !exists {
				continue
			}
			childStat.total++

			if len(childStat.Errors) > 0 {
				t, _ := suggestType(entry)

				if !slices.Contains(childStat.observedTypes, t) {
					childStat.observedTypes = append(childStat.observedTypes, t)
				}
			}

			if fieldVal.Kind() != reflect.Struct && fieldVal.Kind() != reflect.Map {
				if isZero(reflect.ValueOf(entry)) && field.Type.Kind() != reflect.Pointer {
					childStat.unsafeDefaultValue++
				}
			}

			if exists {
				childStat.Compare(entry, fieldVal.Interface())
			}
		}

		f.evalOrAddAsMissing(usedKeys, minimumTypeMap)
	default:
		return
	}
}
