// Package seal provides data sealing and integrity verification.
package seal

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aawadallak/go-core-kit/core/repository"
	"github.com/aawadallak/go-core-kit/pkg/common"
)

type SealedMessage struct {
	common.Entity
	Payload   json.RawMessage
	Signature string `gorm:"index"`
	Secret    string
	Nonce     int64
}

type SealedMessageRepository = repository.AbstractRepository[SealedMessage]

type SealInput struct {
	ExternalID string
	Payload    json.RawMessage
	ExpiresIn  time.Duration
}

type SealOutput struct {
	Signature string
}

type UnsealInput struct {
	Signature string
}

type UnsealOutput struct {
	Payload    json.RawMessage
	ExternalID string
}

type Sealer interface {
	Seal(ctx context.Context, in *SealInput) (*SealOutput, error)
	Unseal(ctx context.Context, in *UnsealInput) (*UnsealOutput, error)
}

type TypeErrUsedSignature struct {
	*common.BaseError
}

func (e *TypeErrUsedSignature) Is(target error) bool {
	_, ok := target.(*TypeErrUsedSignature)
	return ok
}

var ErrUsedSignature = &TypeErrUsedSignature{
	BaseError: &common.BaseError{
		Code:    "SEAL_ERROR",
		Message: "seal error",
	},
}

func NewErrUsedSignature() error {
	return &TypeErrUsedSignature{
		BaseError: &common.BaseError{
			Code:    "USED_SIGNATURE",
			Message: "signature was used before",
		},
	}
}

type TypeErrSealSignatureExpired struct {
	*common.BaseError
}

func (e *TypeErrSealSignatureExpired) Is(target error) bool {
	_, ok := target.(*TypeErrSealSignatureExpired)
	return ok
}

var ErrSealSignatureExpired = &TypeErrSealSignatureExpired{}

func (e *TypeErrSealSignatureExpired) StatusCode() int {
	return http.StatusGone // 410 Gone for an expired resource.
}

func NewErrSealSignatureExpired() error {
	return &TypeErrSealSignatureExpired{
		BaseError: &common.BaseError{
			Code:    "SEAL_SIGNATURE_EXPIRED",
			Message: "The seal signature has expired",
		},
	}
}
