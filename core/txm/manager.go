// Package txm provides txm functionality.
package txm

import "context"

type Fn func(ctx context.Context) error

type Manager interface {
	WithinTransaction(ctx context.Context, fn Fn) error
}
