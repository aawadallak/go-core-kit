package ptr

func Safe[T any](v *T) T {
	if v != nil {
		return *v
	}

	var zero T
	return zero
}
