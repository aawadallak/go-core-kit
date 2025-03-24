package sqs

import (
	"strconv"

	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

func mapToSendMessageInput(encoder broker.Encoder, messages ...broker.Message) (*sqs.SendMessageBatchInput, error) {
	queue := messages[0].Attributes().Get(MessageQueue)

	entries, err := mapToSendMessageEntries(encoder, messages...)
	if err != nil {
		return nil, err
	}

	return &sqs.SendMessageBatchInput{
		Entries:  entries,
		QueueUrl: aws.String(queue),
	}, nil
}

func mapToSendMessageEntries(encoder broker.Encoder, messages ...broker.Message) ([]types.SendMessageBatchRequestEntry, error) {
	entries := make([]types.SendMessageBatchRequestEntry, 0, len(messages))
	for _, message := range messages {
		payload, err := encoder(message)
		if err != nil {
			return nil, err
		}

		entry, err := mapToSendMessageEntry(string(payload), message.Attributes())
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func mapToSendMessageEntry(body string, attributes broker.Attributes) (types.SendMessageBatchRequestEntry, error) {
	res := types.SendMessageBatchRequestEntry{
		Id:                aws.String(uuid.NewString()),
		MessageBody:       aws.String(body),
		MessageAttributes: make(map[string]types.MessageAttributeValue),
	}

	for k, val := range attributes.Values() {
		for _, v := range val {
			res.MessageAttributes[k] = types.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(v),
			}
		}
	}

	if delay, ok := attributes.Lookup(MessageDelaySecond); ok {
		val, err := strconv.ParseInt(delay, 10, 32)
		if err != nil {
			return res, err
		}

		res.DelaySeconds = int32(val)
	}

	return res, nil
}
