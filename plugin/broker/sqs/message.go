package sqs

import (
	"strconv"
	"time"

	"github.com/aawadallak/go-core-kit/core/broker"
)

const (
	MessageIdempotencyKey = "X-Message-Idempotency-Key"
	MessageReceiptHandle  = "X-Message-Receipt-Handle"
	MessageID             = "X-Message-ID"
	MessageQueue          = "X-Message-Queue"
	MessageDelaySecond    = "X-Message-Delay-Second"
)

type message struct {
	data      any
	attr      *attributes
	createdAt time.Time
}

type MessageOption func(*message)

func WithDelaySecond(delay int) MessageOption {
	return func(m *message) {
		m.attr.Add(MessageDelaySecond, strconv.Itoa(delay))
	}
}

func (m *message) Attributes() broker.Attributes { return m.attr }
func (m *message) Payload() any                  { return m.data }

func NewMessage(payload any, queue string, opts ...MessageOption) broker.Message {
	message := &message{
		data:      payload,
		attr:      newAttributes(),
		createdAt: time.Now(),
	}

	for _, opt := range opts {
		opt(message)
	}

	message.attr.Add(MessageQueue, queue)

	return message
}
