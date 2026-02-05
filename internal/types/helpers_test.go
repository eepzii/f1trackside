package types

func valToPtr[T any](val T) *T {
	return &val
}
