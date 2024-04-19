package utils

// simple function allows to convert type to its pointer
func PtrFromVal[T any](value T) *T {
	return &value
}
