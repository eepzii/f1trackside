package scrutineer

import (
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
)

// TODO: add schema validation as well
func (s *TypeStats) Analyze(minimumType any, responseTypeVal any) {
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

	if v.Kind() == reflect.Map {
		iter := v.MapRange()

		for iter.Next() {
			key := iter.Key().String()
			val := iter.Value()

			if entry, exists := minimumTypeMap[key]; exists {
				s.Analyze(entry, val.Interface())
			}
		}
	} else if v.Kind() == reflect.Struct {
		t := v.Type()
		usedKeys := []string{}

		for i := range t.NumField() {
			field := t.Field(i)
			fieldVal := v.Field(i)

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}
			key := strings.Split(jsonTag, ",")[0]
			usedKeys = append(usedKeys, key)

			if s.Children[key] == nil {
				s.Children[key] = &TypeStats{
					Name:     key,
					Children: make(map[string]*TypeStats),
				}
			}
			childStat := s.Children[key]

			entry, exists := minimumTypeMap[key]
			if !exists {
				continue
			}

			if fieldVal.Kind() != reflect.Struct && fieldVal.Kind() != reflect.Map {

				if exists && isZero(fieldVal) {
					childStat.UnsafeDefaultValue++
				}
			}

			if exists {
				childStat.Analyze(entry, fieldVal.Interface())
			}

		}

		for minimumKey := range minimumTypeMap {
			if !slices.Contains(usedKeys, minimumKey) {
				s.Children[minimumKey] = &TypeStats{
					Name:      minimumKey,
					IsMissing: true,
					Children:  make(map[string]*TypeStats),
				}
			}
		}
	}

}

func (s *TypeStats) Tree() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s\n", s.Name)

	s.buildTree(&sb, "")

	return sb.String()
}

func (s *TypeStats) buildTree(sb *strings.Builder, prefix string) {
	keys := make([]string, 0, len(s.Children))
	for k := range s.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		child := s.Children[key]

		isLast := i == len(keys)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}

		msg := "OK"
		if child.IsMissing {
			msg = "MISSING: field found in JSON data but not defined in Go struct"
		} else if child.UnsafeDefaultValue > 0 {
			msg = fmt.Sprintf("FAIL: found %d explicit default values -> suggest pointer",
				child.UnsafeDefaultValue)
		}

		fmt.Fprintf(sb, "%s%s%s  %s\n", prefix, connector, child.Name, msg)

		child.buildTree(sb, childPrefix)
	}
}
