package tool

func ptr[T any](v T) *T {
	return &v
}
