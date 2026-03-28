// Package eventbroker provides event consumption and dispatching via message brokers.
package eventbroker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
)

type HandlerFunc func(ctx context.Context, envelope *Envelope) error

type ConsumerConfig struct {
	EventName     string
	Handler       HandlerFunc
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
	subscription Subscription
}

func NewConsumer(transport ConsumerTransport, cfg ConsumerConfig) (*Consumer, error) { //nolint:gocritic // hugeParam
	if cfg.Handler == nil {
		return nil, errors.New("consumer handler is required")
	}

	cfg.DurableName = normalizeConsumerName(cfg.DurableName)

	// rawHandler deserialises the envelope and delegates to the configured handler.
	rawHandler := func(ctx context.Context, data []byte) error {
		var envelope Envelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			return fmt.Errorf("eventbroker: unmarshal envelope: %w", err)
		}

		if cfg.EventName != "" && envelope.EventName != cfg.EventName {
			// Different event type: skip.
			return nil
		}

		if err := cfg.Handler(ctx, &envelope); err != nil {
			return fmt.Errorf("event %s handler failed: %w", envelope.EventName, err)
		}
		return nil
	}

	subCfg := ConsumerSubscriptionConfig{
		StreamName:    cfg.StreamName,
		Subject:       cfg.Subject,
		DurableName:   cfg.DurableName,
		DLQStreamName: cfg.DLQStreamName,
		DLQSubject:    cfg.DLQSubject,
		MaxDeliver:    cfg.MaxDeliver,
	}

	if cfg.AckWait > 0 {
		subCfg.AckWait = int(cfg.AckWait.Seconds())
	}
	if cfg.FetchMaxWait > 0 {
		subCfg.FetchMaxWait = int(cfg.FetchMaxWait.Seconds())
	}

	sub, err := transport.Subscribe(context.Background(), subCfg, rawHandler)
	if err != nil {
		return nil, err
	}

	return &Consumer{subscription: sub}, nil
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
	c.subscription.Start(ctx)
}

func (c *Consumer) Close() error {
	if c.subscription == nil {
		return nil
	}
	return c.subscription.Close()
}
