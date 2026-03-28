// Package identity provides request context identity extraction.
package identity

import (
	"context"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")

type CtxKey struct{}

type Organization struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Role     string         `json:"role"`
	Metadata map[string]any `json:"metadata"`
}

type Entity struct {
	ExternalID    string         `json:"externalId"`
	FirstName     string         `json:"firstName"`
	LastName      string         `json:"lastName"`
	Email         string         `json:"email"`
	EmailVerified bool           `json:"emailVerified"`
	Picture       string         `json:"picture"`
	Organization  Organization   `json:"organization"`
	Metadata      map[string]any `json:"metadata"`
	Scopes        []string       `json:"scopes"`
}

func (e *Entity) HasOrganization(orgID, role string) error {
	if e.Organization.ID == orgID && e.Organization.Role == role {
		return nil
	}

	return fmt.Errorf("the resource requires the following organization %s with role %s", orgID, role)
}

func (e *Entity) HasOrganizationID(orgID string) error {
	if e.Organization.ID == orgID {
		return nil
	}

	return fmt.Errorf("the resource requires the following organization %s", orgID)
}

func OfContext(ctx context.Context, entity *Entity) context.Context {
	return context.WithValue(ctx, CtxKey{}, entity)
}

func FromContext(ctx context.Context) (*Entity, error) {
	val, ok := ctx.Value(CtxKey{}).(*Entity)
	if !ok {
		return nil, ErrUserNotFound
	}

	return val, nil
}
