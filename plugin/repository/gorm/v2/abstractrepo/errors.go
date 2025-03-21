package abstractrepo

import "errors"

var (
	ErrInvalidType   = errors.New("invalid type")
	ErrInvalidTxType = errors.New("invalid transaction type")
)
