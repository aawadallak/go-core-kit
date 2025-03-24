package sns

import (
	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
)

func mapToPublishInput(encoder broker.Encoder, messages ...broker.Message) (*sns.PublishBatchInput, error) {
	topic := messages[0].Attributes().Get(MessageTopic)

	entries, err := mapToPublishEntries(encoder, messages...)
	if err != nil {
		return nil, err
	}

	return &sns.PublishBatchInput{
		TopicArn:                   aws.String(topic),
		PublishBatchRequestEntries: entries,
	}, nil
}

func mapToPublishEntries(encoder broker.Encoder, messages ...broker.Message) ([]types.PublishBatchRequestEntry, error) {
	entries := make([]types.PublishBatchRequestEntry, 0, len(messages))
	for _, message := range messages {
		payload, err := encoder(message)
		if err != nil {
			return nil, err
		}

		entry, err := mapToPublishEntry(string(payload), message.Attributes())
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func mapToPublishEntry(body string, attributes broker.Attributes) (types.PublishBatchRequestEntry, error) {
	res := types.PublishBatchRequestEntry{
		Id:                aws.String(uuid.NewString()),
		Message:           aws.String(body),
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

	return res, nil
}
