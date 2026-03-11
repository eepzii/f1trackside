package schema

type DataType string

const (
	Any     DataType = "any"
	Bool    DataType = "bool"
	Float64 DataType = "float64"
	Int     DataType = "int"
	Slice   DataType = "[]any"
	// be aware that a struct counts also as map[string]any
	Struct DataType = "struct"
	String DataType = "string"

	Unknown        DataType = "UNKNOWN"
	UnknownNumeric DataType = "UNKNOWN NUMERIC"
)
