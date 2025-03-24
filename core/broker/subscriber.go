package broker

import "context"

// Subscriber is an interface for subscribing to a message queue or topic and receiving messages.
type Subscriber interface {
	// Subscribe initiates a subscription to a message queue or topic and returns a single Message.
	// It takes a context for cancellation and timeout control, returning an error if the subscription fails.
	Subscribe(context.Context) (Message, error)

	// Commit acknowledges the successful processing of a Message.
	// It takes a context and the Message to commit, returning an error if the commit operation fails.
	Commit(context.Context, Message) error

	// Close terminates the subscription and releases any associated resources.
	// It takes a context for cancellation and timeout control.
	Close(context.Context)
}
