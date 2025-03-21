package ssm

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/conf"
)

type noopProvider struct {
	conf.Provider
}

var _ conf.Provider = (*noopProvider)(nil)

func (p *noopProvider) Lookup(key string) (string, bool) {
	return "", false
}

func (p *noopProvider) Scan(fn conf.ScanFunc) {}

func (p *noopProvider) Pull(ctx context.Context, others []conf.Provider) error {
	return nil
}
