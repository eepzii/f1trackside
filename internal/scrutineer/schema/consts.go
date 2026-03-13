package schema

type DataType string

const (
	Any     DataType = "any"
	Bool    DataType = "bool"
	Float64 DataType = "float64"
	Int     DataType = "int"
	Slice   DataType = "[]any"
	Object  DataType = "struct/ map[string]any"
	String  DataType = "string"

	Unknown        DataType = "UNKNOWN"
	UnknownNumeric DataType = "UNKNOWN NUMERIC"
)
