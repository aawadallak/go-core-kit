package sqs

import (
	"encoding/json"

	"github.com/aawadallak/go-core-kit/core/broker"
)

type publisherOption struct {
	encoder broker.Encoder
}

type PublisherOption func(*publisherOption)

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
