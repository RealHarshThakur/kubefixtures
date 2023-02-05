package pkg

// PointerTo returns a pointer to the given value.
func PointerTo[T any](v T) *T {
	return &v
}
