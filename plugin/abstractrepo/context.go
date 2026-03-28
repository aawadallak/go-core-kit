package abstractrepo

import (
	"context"

	"gorm.io/gorm"
)

type ctxKeyTx struct{}

// WithTx returns a new context with the provided transaction added.
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKeyTx{}, tx)
}

// FromContext returns the transaction associated with this context, if any.
func FromContext(ctx context.Context) (*gorm.DB, error) {
	tx, ok := ctx.Value(ctxKeyTx{}).(*gorm.DB)
	if !ok {
		return nil, ErrInvalidTxType
	}

	return tx, nil
}
