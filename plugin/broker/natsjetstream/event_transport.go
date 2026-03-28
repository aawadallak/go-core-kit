package natsjetstream

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/aawadallak/go-core-kit/plugin/event/eventbroker"
)

// Compile-time interface checks.
var (
	_ eventbroker.Transport         = (*EventTransport)(nil)
	_ eventbroker.ConsumerTransport = (*EventTransport)(nil)
)

// EventTransportConfig holds the configuration for creating an EventTransport.
type EventTransportConfig struct {
	Endpoint   string
	StreamName string
	Subjects   []string
}

// EventTransport implements eventbroker.Transport and eventbroker.ConsumerTransport
// on top of NATS JetStream.
type EventTransport struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// NewEventTransport creates a new EventTransport that connects to NATS and
// ensures the given stream exists.
func NewEventTransport(ctx context.Context, cfg EventTransportConfig) (*EventTransport, error) {
	conn, err := nats.Connect(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if cfg.StreamName != "" && len(cfg.Subjects) > 0 {
		if err := ensureStream(ctx, js, cfg.StreamName, cfg.Subjects); err != nil {
			conn.Close()
			return nil, err
		}
	}

	return &EventTransport{conn: conn, js: js}, nil
}

// NewEventTransportFromJetStream creates an EventTransport from an existing
// JetStream handle. Useful when the caller already manages the NATS connection.
func NewEventTransportFromJetStream(js jetstream.JetStream) *EventTransport {
	return &EventTransport{js: js}
}

// Publish implements eventbroker.Transport.
func (t *EventTransport) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := t.js.Publish(ctx, subject, data)
	return err
}

// Subscribe implements eventbroker.ConsumerTransport.
func (t *EventTransport) Subscribe(
	ctx context.Context,
	cfg *eventbroker.ConsumerSubscriptionConfig,
	handler func(ctx context.Context, data []byte) error,
) (eventbroker.Subscription, error) {
	maxDeliver := cfg.MaxDeliver
	if maxDeliver <= 0 {
		maxDeliver = 5
	}
	ackWait := 30 * time.Second
	if cfg.AckWait > 0 {
		ackWait = time.Duration(cfg.AckWait) * time.Second
	}
	fetchMaxWait := 1 * time.Second
	if cfg.FetchMaxWait > 0 {
		fetchMaxWait = time.Duration(cfg.FetchMaxWait) * time.Second
	}

	if err := ensureStream(ctx, t.js, cfg.StreamName, []string{cfg.Subject}); err != nil {
		return nil, err
	}
	if cfg.DLQStreamName != "" && cfg.DLQSubject != "" {
		if err := ensureStream(ctx, t.js, cfg.DLQStreamName, []string{cfg.DLQSubject}); err != nil {
			return nil, err
		}
	}

	stream, err := t.js.Stream(ctx, cfg.StreamName)
	if err != nil {
		return nil, err
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       cfg.DurableName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       ackWait,
		MaxDeliver:    maxDeliver,
		FilterSubject: cfg.Subject,
	})
	if err != nil {
		return nil, err
	}

	return &jsSubscription{
		js:           t.js,
		consumer:     consumer,
		handler:      handler,
		fetchMaxWait: fetchMaxWait,
		dlqSubject:   cfg.DLQSubject,
		hasDLQ:       cfg.DLQStreamName != "" && cfg.DLQSubject != "",
		closeCh:      make(chan struct{}),
	}, nil
}

// Close closes the underlying NATS connection, if owned by this transport.
func (t *EventTransport) Close() error {
	if t.conn == nil {
		return nil
	}
	t.conn.Close()
	return nil
}

// jsSubscription implements eventbroker.Subscription for NATS JetStream.
type jsSubscription struct {
	js           jetstream.JetStream
	consumer     jetstream.Consumer
	handler      func(ctx context.Context, data []byte) error
	fetchMaxWait time.Duration
	dlqSubject   string
	hasDLQ       bool
	closed       atomic.Bool
	closeCh      chan struct{}
	wg           sync.WaitGroup
	closeErr     error
	closeMux     sync.Mutex
	closeOnce    sync.Once
}

func (s *jsSubscription) Start(ctx context.Context) {
	for {
		if s.shouldStop(ctx) {
			return
		}

		msgs, err := s.consumer.Fetch(1, jetstream.FetchMaxWait(s.fetchMaxWait))
		if err != nil {
			if s.shouldStop(ctx) {
				return
			}
			continue
		}

		for msg := range msgs.Messages() {
			s.wg.Add(1)
			func(m jetstream.Msg) {
				defer s.wg.Done()
				if err := s.handleMessage(ctx, m); err != nil {
					log.Println(err)
				}
			}(msg)
		}
	}
}

func (s *jsSubscription) shouldStop(ctx context.Context) bool {
	if ctx.Err() != nil {
		return true
	}
	if s.closed.Load() {
		return true
	}
	select {
	case <-s.closeCh:
		return true
	default:
		return false
	}
}

func (s *jsSubscription) handleMessage(ctx context.Context, msg jetstream.Msg) error {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicErr := fmt.Errorf("panic while handling message: %v", recovered)
			log.Println(panicErr)

			if s.hasDLQ {
				if err := publishDLQ(ctx, s.js, s.dlqSubject, msg.Data(), panicErr.Error()); err != nil {
					log.Printf("failed to publish panic to DLQ: %v", err)
				}
				if err := msg.Ack(); err != nil {
					log.Printf("failed to ack message after panic: %v", err)
				}
				return
			}

			if err := msg.Nak(); err != nil {
				log.Printf("failed to nak message after panic: %v", err)
			}
		}
	}()

	if err := s.handler(ctx, msg.Data()); err != nil {
		if s.hasDLQ {
			if dlqErr := publishDLQ(ctx, s.js, s.dlqSubject, msg.Data(), err.Error()); dlqErr != nil {
				log.Printf("failed to publish handler error to DLQ: %v", dlqErr)
			}
			return msg.Ack()
		}
		return msg.Nak()
	}

	return msg.Ack()
}

func (s *jsSubscription) Close() error {
	s.closeOnce.Do(func() {
		s.closed.Store(true)
		close(s.closeCh)
		s.wg.Wait()
	})

	s.closeMux.Lock()
	defer s.closeMux.Unlock()
	return s.closeErr
}
