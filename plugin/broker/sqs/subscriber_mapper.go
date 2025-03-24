package sqs

import (
	"time"

	"github.com/aawadallak/go-core-kit/core/broker"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func mapToReceiveMessageInput(queue string) *sqs.ReceiveMessageInput {
	return &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queue),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
	}
}

func mapToCommitMessageInput(queue string, message broker.Message) *sqs.DeleteMessageInput {
	return &sqs.DeleteMessageInput{
		QueueUrl: aws.String(queue),
		ReceiptHandle: aws.String(
			message.Attributes().Get(MessageReceiptHandle),
		),
	}
}

func mapProviderToMessage(decoder broker.Decoder, queue string, m types.Message) (broker.Message, error) {
	body, err := decoder([]byte(*m.Body))
	if err != nil {
		return nil, err
	}

	res := &message{
		data:      body,
		createdAt: time.Now(),
		attr:      newAttributes(),
	}

	res.Attributes().Add(MessageID, *m.MessageId)
	res.Attributes().Add(MessageIdempotencyKey, m.Attributes["MessageDeduplicationId"])
	res.Attributes().Add(MessageReceiptHandle, *m.ReceiptHandle)
	res.Attributes().Add(MessageQueue, queue)

	return res, nil
}
