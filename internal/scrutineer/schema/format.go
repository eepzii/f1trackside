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

		child.WriteTree(sb, childPrefix)
	}
}

func (f *Field) buildMessage(indent string) string {
	if f.unsafeDefaultValue > 0 {
		return fmt.Sprintf("FAIL: found %d explicit default values -> suggest pointer",
			f.unsafeDefaultValue)
	}

	var msg = "OK"

	var suggestedTypes string
	for _, t := range f.observedTypes {
		if suggestedTypes != "" {
			suggestedTypes += fmt.Sprintf(", %s", t)
			continue
		}
		suggestedTypes += t
	}

	var typeErrors strings.Builder
	if len(f.Errors) > 0 {
		for err := range f.Errors {
			fmt.Fprintf(&typeErrors, "\n%s%s", indent, err)
		}
	}

	if slices.Contains(f.observedTypes, "struct") &&
		slices.Contains(f.observedTypes, "[]any") &&
		(len(f.observedTypes) == 2 || len(f.Errors) == 2) {
		suggestedTypes = fmt.Sprintf("dynamic JSON array (%s)", suggestedTypes)
	}

	if f.isMissing {
		msg = fmt.Sprintf("MISSING: field found in JSON data but not defined in Go struct -> suggest adding as %s",
			suggestedTypes)
	} else if len(f.Errors) > 0 {
		msg = fmt.Sprintf("TYPE ERROR: suggest %s%s",
			suggestedTypes, typeErrors.String())
	}

	return msg
}
