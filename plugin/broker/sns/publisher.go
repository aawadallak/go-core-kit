package sns

import (
	"context"
	"fmt"

	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type Publisher struct {
	options  *publisherOption
	provider *sns.Client
}

var _ broker.Publisher = (*Publisher)(nil)

func NewPublisher(ctx context.Context, opts ...PublisherOption) (*Publisher, error) {
	options := newPublisherOption(opts...)

	if options.client == nil {
		provider, err := newClient(ctx)
		if err != nil {
			return nil, err
		}

		options.client = provider
	}

	return &Publisher{
		options:  options,
		provider: options.client,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, messages ...broker.Message) error {
	if len(messages) > 10 {
		return ErrSizeLimit
	}

	req, err := mapToPublishInput(p.options.encoder, messages...)
	if err != nil {
		return err
	}

	result, err := p.provider.PublishBatch(ctx, req)
	if err != nil {
		return err
	}

	for _, r := range result.Failed {
		return fmt.Errorf("failed to publish message: %s", *r.Message)
	}

	return nil
}
