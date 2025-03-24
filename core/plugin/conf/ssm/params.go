package ssm

import (
	"context"
	"fmt"
	"strings"

	"github.com/aawadallak/go-core-kit/core/conf"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// ssmParameter represents a reference to an AWS SSM parameter.
// It contains both the configuration key and the SSM parameter name.
type ssmParameter struct {
	key   string
	value string
}

// findSSMParameterReferences scans the provided configuration providers for SSM parameter references.
// It looks for values that start with the SSM prefix pattern and extracts the parameter references.
// Returns a slice of ssmParameter containing the key-value pairs of found SSM references.
func findSSMParameterReferences(providers []conf.Provider) []ssmParameter {
	var parameters []ssmParameter

	for _, provider := range providers {
		provider.Scan(func(key, value string) {
			if !strings.HasPrefix(value, prefixPattern) {
				return
			}

			parameters = append(parameters, ssmParameter{
				key:   key,
				value: strings.Split(value, prefixPattern)[1],
			})
		})
	}

	return parameters
}

// fetchSSMParameters retrieves parameter values from AWS SSM using the provided parameter names.
// It fetches the parameters with decryption enabled and stores them in the provider's data map.
// Returns an error if the parameters cannot be retrieved or if any parameters are invalid.
func (s *provider) fetchSSMParameters(ctx context.Context, parameterNames []string) error {
	res, err := s.client.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          parameterNames,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to get parameters: %w", err)
	}

	if len(res.InvalidParameters) > 0 {
		return fmt.Errorf("invalid parameters: %v", res.InvalidParameters)
	}

	for _, param := range res.Parameters {
		s.data[*param.Name] = *param.Value
	}

	return nil
}
