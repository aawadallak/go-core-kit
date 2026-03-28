// Package bcrypt provides bcrypt functionality.
package bcrypt

import (
	"errors"

	"github.com/aawadallak/go-core-kit/core/cipher"

	"golang.org/x/crypto/bcrypt"
)

type Adapter struct{}

func NewAdapter() cipher.Cipher {
	return &Adapter{}
}

func (a *Adapter) Encrypt(value string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
}

func (a *Adapter) Verify(hashedValue []byte, value string) error {
	if err := bcrypt.CompareHashAndPassword(hashedValue, []byte(value)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return cipher.NewErrInvalidHash()
		}

		return err
	}

	return nil
}
