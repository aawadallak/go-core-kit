package sqs

import (
	"context"
	"errors"

	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var ErrNoMessageInQueue = errors.New("no message in queue")

type Subscriber struct {
	options  *subscriberOptions
	provider *sqs.Client
	queue    string
}

var _ broker.Subscriber = (*Subscriber)(nil)

func (s *Subscriber) Subscribe(ctx context.Context) (broker.Message, error) {
	result, err := s.provider.ReceiveMessage(ctx, mapToReceiveMessageInput(s.queue))
	if err != nil {
		return nil, err
	}

	if len(result.Messages) == 0 {
		return nil, ErrNoMessageInQueue
	}

	return mapProviderToMessage(
		s.options.decoder, s.queue, result.Messages[0])
}

func (s *Subscriber) Commit(ctx context.Context, message broker.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	input := mapToCommitMessageInput(s.queue, message)
	_, err := s.provider.DeleteMessage(ctx, input)
	if err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) Close(ctx context.Context) {}

func NewSubscriber(ctx context.Context, queue string, opts ...SubscriberOption) (*Subscriber, error) {
	options := newSubscriberOption(opts...)

	if options.client == nil {
		provider, err := newClient(ctx)
		if err != nil {
			return nil, err
		}

		options.client = provider
	}

	subscriber := &Subscriber{
		options:  options,
		queue:    queue,
		provider: options.client,
	}

	return subscriber, nil
}
