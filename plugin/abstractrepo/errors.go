package abstractrepo

import "errors"

var (
	// ErrInvalidType is returned when the type is invalid.
	ErrInvalidType = errors.New("invalid type")
	// ErrInvalidTxType is returned when the context key is invalid.
	ErrInvalidTxType = errors.New("invalid transaction type")
)
