// Package ptr provides ptr functionality.
package ptr

func New[T any](value T) *T {
	return &value
}
