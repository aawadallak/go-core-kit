// Package onepassword provides a 1Password configuration provider for core/conf.
package onepassword

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/1password/onepassword-sdk-go"
	"github.com/aawadallak/go-core-kit/core/conf"
)

const opPrefix = "op://"

type provider struct {
	data   map[string]string
	client *onepassword.Client
}

var _ conf.Provider = (*provider)(nil)

// NewProvider creates a 1Password configuration provider.
// Returns a noop provider if OP_SERVICE_ACCOUNT_TOKEN is not set or client creation fails.
func NewProvider(ctx context.Context) conf.Provider {
	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	if token == "" {
		return &noopProvider{}
	}
	client, err := onepassword.NewClient(ctx,
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("go-core-kit", "v1.0.0"),
	)
	if err != nil {
		return &noopProvider{}
	}
	return &provider{client: client, data: make(map[string]string)}
}

func (p *provider) Lookup(key string) (string, bool) {
	v, ok := p.data[key]
	return v, ok
}

func (p *provider) Scan(fn conf.ScanFunc) {
	for k, v := range p.data {
		fn(k, v)
	}
}

func (p *provider) Load(ctx context.Context, others []conf.Provider) error {
	refs := findReferences(others)
	var errs []error
	for _, ref := range refs {
		secret, err := p.client.Secrets().Resolve(ctx, ref.opRef)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to resolve 1password secret %s: %w", ref.opRef, err))
			continue
		}
		p.data[ref.key] = secret
	}
	return errors.Join(errs...)
}

type opRef struct {
	key   string // config key (e.g., "DB_PASS")
	opRef string // full 1password reference (e.g., "op://vault/item/field")
}

func findReferences(others []conf.Provider) []opRef {
	var refs []opRef
	for _, other := range others {
		other.Scan(func(key, value string) {
			if strings.HasPrefix(value, opPrefix) {
				refs = append(refs, opRef{key: key, opRef: value})
			}
		})
	}
	return refs
}
