package schema

func (f *Field) AddByPath(path []string) *Field {
	current := f

	for _, fieldName := range path {
		if current.Children[fieldName] == nil {
			current.Children[fieldName] = &Field{
				Name:     fieldName,
				Errors:   make(map[string]int),
				Children: make(map[string]*Field),
			}
		}
		current = current.Children[fieldName]
	}

	return current
}
