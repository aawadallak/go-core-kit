package sns

import (
	"time"

	"github.com/aawadallak/go-core-kit/core/broker"
)

const MessageTopic = "topic"

type message struct {
	data      any
	attr      *attributes
	createdAt time.Time
}

type MessageOption func(*message)

func (m *message) Attributes() broker.Attributes { return m.attr }
func (m *message) Payload() any                  { return m.data }

func NewMessage(payload any, topic string, opts ...MessageOption) broker.Message {
	message := &message{
		data:      payload,
		attr:      newAttributes(),
		createdAt: time.Now(),
	}

	message.attr.Add(MessageTopic, topic)
	for _, opt := range opts {
		opt(message)
	}

	return message
}
