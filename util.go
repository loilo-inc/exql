package exql

// Ptr returns the pointer of the argument.
func Ptr[T any](t T) *T {
	return &t
}
