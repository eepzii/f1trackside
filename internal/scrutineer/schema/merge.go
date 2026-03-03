package schema

import "slices"

func (f *Field) Merge(other *Field) {
	f.total += other.total
	f.unsafeDefaultValue += other.unsafeDefaultValue
	if other.isMissing {
		f.isMissing = true
	}

	if f.Errors == nil && other.Errors != nil {
		f.Errors = make(map[string]int)
	}
	for err, occurrence := range other.Errors {
		f.Errors[err] += occurrence
	}

	for _, t := range other.observedTypes {
		if !slices.Contains(f.observedTypes, t) {
			f.observedTypes = append(f.observedTypes, t)
		}
	}

	if f.Children == nil && other.Children != nil {
		f.Children = make(map[string]*Field)
	}
	for name, otherChild := range other.Children {
		if fieldChild, exists := f.Children[name]; exists {
			fieldChild.Merge(otherChild)
			continue
		}
		f.Children[name] = otherChild
	}
}
