package schema

type Field struct {
	Name     string
	Errors   map[string]int
	Children map[string]*Field

	isMissing          bool
	total              int
	unsafeDefaultValue int
	observedTypes      []DataType
}
