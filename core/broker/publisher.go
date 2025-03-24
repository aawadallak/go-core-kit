package broker

import "context"

// Publisher is an interface for publishing messages to a message queue or topic.
type Publisher interface {
	// Publish publishes a message to a message queue or topic.
	Publish(context.Context, ...Message) error
}
