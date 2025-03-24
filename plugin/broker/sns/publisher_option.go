package sns

import (
	"encoding/json"

	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type publisherOption struct {
	encoder broker.Encoder
	client  *sns.Client
}

type PublisherOption func(*publisherOption)

func WithAwsClient(client *sns.Client) PublisherOption {
	return func(o *publisherOption) {
		o.client = client
	}
}

func newPublisherOption(opts ...PublisherOption) *publisherOption {
	options := &publisherOption{
		encoder: func(m broker.Message) ([]byte, error) {
			return json.Marshal(m.Payload())
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
