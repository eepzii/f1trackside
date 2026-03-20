package schema

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

func (f *Field) WriteTree(sb *strings.Builder, prefix string) {
	keys := make([]string, 0, len(f.Children))
	for k := range f.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		child := f.Children[key]

		isLast := i == len(keys)-1

		connector := "├── "
		if isLast {
			connector = "└── "
		}

		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}

		msg := child.buildMessage(childPrefix)

		fmt.Fprintf(sb, "%s%s%s  %s\n",
			prefix, connector, child.Name, msg)

		if slices.Contains(child.observedTypes, Slice) && len(child.observedTypes) == 1 {
			continue
		}
		child.WriteTree(sb, childPrefix)
	}
}

func (f *Field) buildMessage(indent string) string {
	var msg = "OK"

	var suggestedTypes DataType
	for _, t := range f.observedTypes {
		if suggestedTypes != "" {
			suggestedTypes += DataType(fmt.Sprintf(", %s", t))
			continue
		}
		suggestedTypes += t
	}

	var typeErrors strings.Builder
	if len(f.Errors) > 0 {
		for err := range f.Errors {
			if err == "empty struct" {
				continue
			}
			fmt.Fprintf(&typeErrors, "\n%s%s", indent, err)
		}
	}

	if slices.Contains(f.observedTypes, Object) &&
		slices.Contains(f.observedTypes, Slice) &&
		(len(f.observedTypes) == 2 || len(f.Errors) == 2) {
		suggestedTypes = DataType(fmt.Sprintf("dynamic JSON array (%s)", suggestedTypes))
	}

	if f.isMissing {
		msg = fmt.Sprintf("MISSING: field found in JSON data but not defined in Go struct -> suggest adding as %s",
			suggestedTypes)
	} else if len(f.Errors) > 0 {
		var arrayTypes strings.Builder
		var foundTypes []DataType
		if slices.Contains(f.observedTypes, Slice) && len(f.observedTypes) == 1 {
			for _, child := range f.Children {
				for _, observedType := range child.observedTypes {
					if !slices.Contains(foundTypes, observedType) {
						foundTypes = append(foundTypes, observedType)
					}
				}
			}
			var types strings.Builder
			for i, t := range foundTypes {
				if i == 0 {
					fmt.Fprint(&types, t)
					continue
				}
				fmt.Fprintf(&types, ", %s", t)
			}
			fmt.Fprintf(&arrayTypes, "\n%sArray has item(s) with types of: \"%s\"", indent, types.String())
			if len(foundTypes) == 1 {
				fmt.Fprintf(&arrayTypes, "\n%s=> consider using []%s instead of suggested []any", indent, foundTypes[0])
			}

			typeErrors = strings.Builder{}
		}

		msg = fmt.Sprintf("TYPE ERROR: suggest %s%s%s",
			suggestedTypes, typeErrors.String(), arrayTypes.String())
	} else if f.unsafeDefaultValue > 0 {
		return fmt.Sprintf("FAIL: found %d explicit default values -> suggest pointer",
			f.unsafeDefaultValue)
	}

	return msg
}
