package eventbroker

import "context"

// Transport publishes serialized event envelopes to a subject/topic.
type Transport interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

// ConsumerTransport abstracts the consumption side so that eventbroker
// does not depend on a concrete broker implementation.
type ConsumerTransport interface {
	// Subscribe creates a subscription that delivers raw messages to the handler.
	// The handler receives the raw bytes; deserialization is done by the caller.
	Subscribe(ctx context.Context, cfg ConsumerSubscriptionConfig, handler func(ctx context.Context, data []byte) error) (Subscription, error)
}

// ConsumerSubscriptionConfig carries the parameters needed to create a subscription.
type ConsumerSubscriptionConfig struct {
	StreamName    string
	Subject       string
	DurableName   string
	DLQStreamName string
	DLQSubject    string
	MaxDeliver    int
	AckWait       int // seconds
	FetchMaxWait  int // seconds
}

// Subscription represents an active subscription that can be started and closed.
type Subscription interface {
	Start(ctx context.Context)
	Close() error
}
