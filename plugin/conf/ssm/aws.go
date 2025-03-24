package ssm

import (
	"context"
	"fmt"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// newSSMClient creates a new AWS SSM client using default AWS configuration.
// It loads the AWS configuration from the environment and returns a configured SSM client.
// Returns an error if the AWS configuration cannot be loaded.
func newSSMClient(ctx context.Context) (*ssm.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return ssm.NewFromConfig(cfg), nil
}
