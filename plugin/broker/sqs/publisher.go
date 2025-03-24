package sqs

import (
	"context"
	"errors"

	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Publisher struct {
	options  *publisherOption
	provider *sqs.Client
}

var _ broker.Publisher = (*Publisher)(nil)

func (p *Publisher) Publish(ctx context.Context, messages ...broker.Message) error {
	if len(messages) > 10 {
		return ErrSizeLimit
	}

	req, err := mapToSendMessageInput(p.options.encoder, messages...)
	if err != nil {
		return err
	}

	result, err := p.provider.SendMessageBatch(ctx, req)
	if err != nil {
		return err
	}

	var publishErr []error
	for _, r := range result.Failed {
		publishErr = append(publishErr, newPublishMessageError(r))
	}

	return errors.Join(publishErr...)
}

func NewPublisher(ctx context.Context, opts ...PublisherOption) (*Publisher, error) {
	options := newPublisherOption(opts...)

	provider, err := newClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		options:  options,
		provider: provider,
	}, nil
}

func newPublishMessageError(cause types.BatchResultErrorEntry) error {
	return &ErrSendMessage{
		Code:    *cause.Code,
		Message: *cause.Message,
	}
}
