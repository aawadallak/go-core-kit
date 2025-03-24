package ssm

import (
	"context"
	"fmt"

	"github.com/aawadallak/go-core-kit/core/conf"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	prefixPattern       = "ssm://"
	maxParamsPerRequest = 10
)

type provider struct {
	data   map[string]string
	client *ssm.Client
}

var _ conf.Provider = (*provider)(nil)

// NewProvider creates a new SSM configuration provider.
// It initializes an AWS SSM client and returns a provider that can fetch parameters from AWS SSM.
// If the SSM client initialization fails, it returns a no-op provider that does nothing.
func NewProvider() conf.Provider {
	client, err := newSSMClient(context.TODO())
	if err != nil {
		return &noopProvider{}
	}

	return &provider{
		client: client,
		data:   make(map[string]string),
	}
}

// Lookup retrieves a value by key from the SSM source
func (p *provider) Lookup(key string) (string, bool) {
	v, ok := p.data[key]
	return v, ok
}

// Scan iterates over all key-value pairs in the SSM source
func (p *provider) Scan(fn conf.ScanFunc) {
	for k, v := range p.data {
		fn(k, v)
	}
}

// Load fetches parameters from SSM and updates the local data store
func (p *provider) Load(ctx context.Context, others []conf.Provider) error {
	parameters := findSSMParameterReferences(others)
	if len(parameters) == 0 {
		return nil
	}

	// Process secrets in batches
	for i := 0; i < len(parameters); i += maxParamsPerRequest {
		end := i + maxParamsPerRequest
		if end > len(parameters) {
			end = len(parameters)
		}

		batch := parameters[i:end]
		params := make([]string, len(batch))
		for j, param := range batch {
			params[j] = param.value
		}

		if err := p.fetchSSMParameters(ctx, params); err != nil {
			return fmt.Errorf("failed to load parameters batch: %w", err)
		}
	}

	// Update the data store with fetched values
	for _, param := range parameters {
		plainText, ok := p.data[param.value]
		if !ok {
			return fmt.Errorf("parameter not found: %s=%s", param.key, param.value)
		}

		p.data[param.key] = plainText
		p.data[param.value] = plainText
	}

	return nil
}
