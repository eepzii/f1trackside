package schema

import (
	"reflect"
	"slices"
)

func (f *Field) checkMissingFields(usedFields []string, minimumTypeMap map[string]any) {
	for minimumKey := range minimumTypeMap {
		if !slices.Contains(usedFields, minimumKey) {
			suggestedType := suggestJSONType(minimumTypeMap[minimumKey])

			if f.Children[minimumKey] == nil {
				f.Children[minimumKey] = &Field{
					Name:      minimumKey,
					isMissing: true,
					Children:  make(map[string]*Field),
				}
			}

			childStat := f.Children[minimumKey]
			childStat.total++

			if !slices.Contains(childStat.observedTypes, suggestedType) {
				childStat.observedTypes = append(childStat.observedTypes, suggestedType)
			}
			f.Children[minimumKey] = childStat
		}
	}
}

func (f *Field) checkUnsafeDefaults(fieldType reflect.Type, fieldVal reflect.Value, entry any) {
	if fieldType.Kind() == reflect.Struct && fieldType.NumField() == 0 {
		f.Errors["empty struct"]++
		return
	}

	var entryVal = reflect.ValueOf(entry)

	if entryVal.Kind() == reflect.Slice || entryVal.Kind() == reflect.Array {
		if fieldType.Kind() != reflect.Slice && fieldType.Kind() != reflect.Array && fieldType.Kind() != reflect.Map {
			return
		}

		for _, elem := range entry.([]any) {
			checkVal, err := normalizeJSONNumber(elem)
			if err != nil {
				f.Errors[err.Error()]++
				continue
			}

			if isZero(reflect.ValueOf(checkVal)) && fieldType.Elem().Kind() != reflect.Pointer {
				f.unsafeDefaultValue++
			}
		}
	} else if entryVal.Kind() == reflect.Map {
		if fieldType.Kind() != reflect.Map {
			return
		}

		for _, elem := range entry.(map[string]any) {
			checkVal, err := normalizeJSONNumber(elem)
			if err != nil {
				f.Errors[err.Error()]++
				continue
			}

			if isZero(reflect.ValueOf(checkVal)) && fieldType.Elem().Kind() != reflect.Pointer {
				f.unsafeDefaultValue++
			}
		}
	} else if fieldVal.Kind() != reflect.Struct && fieldVal.Kind() != reflect.Map {
		checkVal, err := normalizeJSONNumber(entry)
		if err != nil {
			f.Errors[err.Error()]++
			return
		}

		if isZero(reflect.ValueOf(checkVal)) && fieldType.Kind() != reflect.Pointer {
			f.unsafeDefaultValue++
		}
	}
}
