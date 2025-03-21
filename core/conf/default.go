package conf

import (
	"context"
	"os"
)

// defaultProvider implements the Provider interface using environment variables.
type defaultProvider struct {
	values map[string]string
}

var _ Provider = (*defaultProvider)(nil)

// Lookup retrieves a configuration value from environment variables.
// Returns the value and a boolean indicating if the key was found.
func (d *defaultProvider) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}

// Scan iterates over all key-value pairs in the default provider.
// The provided function is called for each key-value pair.
func (d *defaultProvider) Scan(fn ScanFunc) {
	for key, value := range d.values {
		fn(key, value)
	}
}

// Pull is a no-op for the default provider since environment variables
// are always available at runtime.
func (d *defaultProvider) Pull(ctx context.Context, providers []Provider) error {
	// We don't need to pull because
	// env variables are always available
	return nil
}

// NewProvider creates a new default provider that uses environment variables.
// Returns a Provider implementation.
func NewProvider() Provider {
	return &defaultProvider{}
}
