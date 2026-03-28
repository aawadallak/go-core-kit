// Package seal provides seal functionality.
package seal

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aawadallak/go-core-kit/pkg/common"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
)

type SealerDependencies struct {
	RepoSealedMessage SealedMessageRepository
}

type Sealer struct {
	deps *SealerDependencies
}

var _ SealerInterface = (*Sealer)(nil)

func (s *Sealer) Seal(ctx context.Context, in *SealInput) (*SealOutput, error) {
	entity := &SealedMessage{
		Secret: uuid.NewString(),
	}

	if in.Payload != nil {
		payload, err := json.Marshal(in.Payload)
		if err != nil {
			return nil, common.NewErrInternalServer(err)
		}

		entity.Payload = payload
	}

	expiresIn := time.Minute * 30
	if in.ExpiresIn > 0 {
		expiresIn = in.ExpiresIn
	}

	token, err := jwt.NewBuilder().
		Subject(in.ExternalID).
		Expiration(time.Now().Add(expiresIn)).
		NotBefore(time.Now()).
		IssuedAt(time.Now()).
		Build()
	if err != nil {
		return nil, common.NewErrInternalServer(err)
	}

	bSig, err := jwt.Sign(token, jwa.HS256, []byte(entity.Secret))
	if err != nil {
		return nil, common.NewErrInternalServer(err)
	}

	entity.Signature = string(bSig)
	if _, err := s.deps.RepoSealedMessage.Save(ctx, entity); err != nil {
		return nil, common.NewErrInternalServer(err)
	}

	return &SealOutput{
		Signature: entity.Signature,
	}, nil
}

func (s *Sealer) Unseal(ctx context.Context, in *UnsealInput) (*UnsealOutput, error) {
	sig, err := s.deps.RepoSealedMessage.FindOne(ctx, &SealedMessage{Signature: in.Signature})
	if err != nil {
		if errors.Is(err, common.ErrResourceNotFound) {
			return nil, common.NewErrResourceNotFound(err)
		}

		return nil, common.NewErrInternalServer(err)
	}

	if sig.Nonce != 0 {
		return nil, NewErrUsedSignature()
	}

	sig.Nonce = time.Now().Unix()

	options := []jwt.ParseOption{
		jwt.WithVerify(jwa.HS256, []byte(sig.Secret)),
	}

	token, err := jwt.ParseString(in.Signature, options...)
	if err != nil {
		return nil, common.NewErrInternalServer(err)
	}

	if token.Expiration().Before(time.Now()) {
		return nil, NewErrSealSignatureExpired()
	}

	if _, err := s.deps.RepoSealedMessage.Update(ctx, sig); err != nil {
		return nil, common.NewErrInternalServer(err)
	}

	return &UnsealOutput{
		Payload:    sig.Payload,
		ExternalID: token.Subject(),
	}, nil
}

func New(deps *SealerDependencies) *Sealer {
	return &Sealer{
		deps: common.MustValidateDependencies(deps),
	}
}
