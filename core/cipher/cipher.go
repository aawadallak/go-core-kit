// Package cipher provides cipher functionality.
package cipher

import (
	"github.com/aawadallak/go-core-kit/pkg/common"
)

type TypeErrInvalidHash struct {
	*common.BaseError
}

func (e *TypeErrInvalidHash) Is(target error) bool {
	_, ok := target.(*TypeErrInvalidHash)
	return ok
}

var ErrInvalidHash = &TypeErrInvalidHash{}

func NewErrInvalidHash() error {
	return &TypeErrInvalidHash{
		BaseError: &common.BaseError{
			Code:    "INVALID_HASH",
			Message: "The hash is invalid",
		},
	}
}

type Cipher interface {
	Encrypt(string) ([]byte, error)
	Verify(hashedValue []byte, value string) error
}
