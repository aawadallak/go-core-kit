package nats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn    *nats.Conn
	subject string
}

type PublisherConfig struct {
	Endpoint string
	Subject  string
}

type PublishMessage struct {
	Message any
}

func NewPublisher(cfg PublisherConfig) (*Publisher, error) {
	conn, err := nats.Connect(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		subject: cfg.Subject,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, req PublishMessage) error {
	payload, err := json.Marshal(req.Message)
	if err != nil {
		return err
	}

	if err := p.conn.Publish(p.subject, payload); err != nil {
		return err
	}

	// Force buffered client data to be sent immediately.
	// nats.FlushWithContext requires a context with deadline.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return p.conn.FlushWithContext(timeoutCtx)
	}

	return p.conn.FlushWithContext(ctx)
}

func (p *Publisher) Close() error {
	if p.conn == nil {
		return nil
	}
	p.conn.Close()
	return nil
}
