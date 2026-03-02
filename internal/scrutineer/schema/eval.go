package schema

import (
	"slices"
)

func (f *Field) evalOrAddAsMissing(usedFields []string, minimumTypeMap map[string]any) {
	for minimumKey := range minimumTypeMap {
		if !slices.Contains(usedFields, minimumKey) {
			suggestedType, err := suggestType(minimumTypeMap[minimumKey])
			if err != nil {
				suggestedType = "UNKNOWN"
			}

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
