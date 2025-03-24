package sqs

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type subscriberOptions struct {
	decoder broker.Decoder
	client  *sqs.Client
}

type SubscriberOption func(*subscriberOptions)

func WithDecoder(decoder broker.Decoder) SubscriberOption {
	return func(so *subscriberOptions) {
		so.decoder = decoder
	}
}

func WithDecoderTarget(typeof any) SubscriberOption {
	return func(s *subscriberOptions) {
		v := reflect.TypeOf(typeof)

		s.decoder = func(b []byte) (any, error) {
			target := reflect.New(v).Interface()
			if err := json.Unmarshal(b, target); err != nil {
				return nil, err
			}

			return target, nil
		}
	}
}

func WithSNSEventDecoder(typeof any) SubscriberOption {
	type snsEvent struct {
		Message          json.RawMessage
		Type             string
		MessageId        string
		TopicArn         string
		Subject          string
		Timestamp        string
		SignatureVersion string
		Signature        string
		SigningCertURL   string
		UnsubscribeURL   string
	}

	return func(s *subscriberOptions) {
		s.decoder = func(b []byte) (any, error) {
			event := &snsEvent{}
			if err := json.Unmarshal(b, event); err != nil {
				return nil, err
			}

			message, err := strconv.Unquote(string(event.Message))
			if err != nil {
				return nil, err
			}

			target := reflect.New(reflect.TypeOf(typeof)).Interface()
			if err := json.Unmarshal([]byte(message), target); err != nil {
				return nil, err
			}

			return target, nil
		}
	}
}

func WithAwsClient(client *sqs.Client) SubscriberOption {
	return func(s *subscriberOptions) {
		s.client = client
	}
}

func newSubscriberOption(opts ...SubscriberOption) *subscriberOptions {
	options := &subscriberOptions{
		decoder: func(b []byte) (any, error) {
			var target any

			if err := json.Unmarshal(b, &target); err != nil {
				return nil, err
			}

			return target, nil
		},
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}
