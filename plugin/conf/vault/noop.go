package vault

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/conf"
)

type noopProvider struct{}

var _ conf.Provider = (*noopProvider)(nil)

func (n *noopProvider) Load(_ context.Context, _ []conf.Provider) error { return nil }
func (n *noopProvider) Lookup(_ string) (string, bool)                  { return "", false }
func (n *noopProvider) Scan(_ conf.ScanFunc)                            {}
