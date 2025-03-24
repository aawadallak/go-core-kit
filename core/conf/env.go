package conf

import (
	"context"
	"os"
	"strings"
)

// envProvider implements the Provider interface using environment variables.
type envProvider struct{}

var _ Provider = (*envProvider)(nil)

// Lookup retrieves a configuration value from environment variables.
// Returns the value and a boolean indicating if the key was found.
func (d *envProvider) Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}

// Scan iterates over all key-value pairs in the env provider.
// The provided function is called for each key-value pair.
func (d *envProvider) Scan(fn ScanFunc) {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		fn(pair[0], pair[1])
	}
}

// Load is a no-op for the env provider
func (d *envProvider) Load(ctx context.Context, _ []Provider) error {
	return nil
}

// newEnvProvider creates a new env provider
// Returns a Provider implementation.
func newEnvProvider() Provider {
	return &envProvider{}
}
