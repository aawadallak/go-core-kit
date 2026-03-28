// Package eventbroker provides event consumption and dispatching via message brokers.
package eventbroker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	brokerjs "github.com/aawadallak/go-core-kit/plugin/broker/natsjetstream"
)

type HandlerFunc func(ctx context.Context, envelope *Envelope) error

type ConsumerConfig struct {
	EventName     string
	Handler       HandlerFunc
	Endpoint      string
	StreamName    string
	Subject       string
	DurableName   string
	DLQStreamName string
	DLQSubject    string
	MaxDeliver    int
	AckWait       time.Duration
	FetchMaxWait  time.Duration
}

type Consumer struct {
	worker *brokerjs.Worker[Envelope]
}

func NewConsumer(ctx context.Context, cfg ConsumerConfig) (*Consumer, error) { //nolint:gocritic // hugeParam
	if cfg.Handler == nil {
		return nil, errors.New("consumer handler is required")
	}

	cfg.DurableName = normalizeConsumerName(cfg.DurableName)

	wrappedHandler := func(ctx context.Context, envelope *Envelope) error {
		if cfg.EventName != "" && envelope.EventName != cfg.EventName {
			// Different event type: ack and skip.
			return nil
		}

		if err := cfg.Handler(ctx, envelope); err != nil {
			return fmt.Errorf("event %s handler failed: %w", envelope.EventName, err)
		}
		return nil
	}

	worker, err := brokerjs.NewWorker(ctx, brokerjs.WorkerConfig[Envelope]{
		Endpoint:      cfg.Endpoint,
		StreamName:    cfg.StreamName,
		Subject:       cfg.Subject,
		DurableName:   cfg.DurableName,
		DLQStreamName: cfg.DLQStreamName,
		DLQSubject:    cfg.DLQSubject,
		MaxDeliver:    cfg.MaxDeliver,
		AckWait:       cfg.AckWait,
		FetchMaxWait:  cfg.FetchMaxWait,
		Handler:       wrappedHandler,
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{worker: worker}, nil
}

// NATS JetStream consumer names are strict (letters, digits, '_' and '-').
// Normalize configured names so env/catalog names like "pixonlive.saas.listener.v1"
// do not crash startup.
func normalizeConsumerName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return name
	}

	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}

	return b.String()
}

func (c *Consumer) Start(ctx context.Context) {
	c.worker.Start(ctx)
}

func (c *Consumer) Close() error {
	if c.worker == nil {
		return nil
	}
	return c.worker.Close()
}
