// Package natsjetstream provides NATS JetStream message publishing and consumption.
package natsjetstream

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type PublisherConfig struct {
	Endpoint   string
	StreamName string
	Subject    string
}

type PublishMessage struct {
	Message any
}

type Publisher struct {
	conn    *nats.Conn
	js      jetstream.JetStream
	subject string
}

func NewPublisher(ctx context.Context, cfg PublisherConfig) (*Publisher, error) {
	conn, err := nats.Connect(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := ensureStream(ctx, js, cfg.StreamName, []string{cfg.Subject}); err != nil {
		conn.Close()
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		js:      js,
		subject: cfg.Subject,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, req PublishMessage) error {
	payload, err := json.Marshal(req.Message)
	if err != nil {
		return err
	}

	_, err = p.js.Publish(ctx, p.subject, payload)
	return err
}

func (p *Publisher) Close() error {
	if p.conn == nil {
		return nil
	}
	p.conn.Close()
	return nil
}

func ensureStream(ctx context.Context, js jetstream.JetStream, name string, subjects []string) error {
	_, err := js.Stream(ctx, name)
	if err == nil {
		return nil
	}

	_, err = js.CreateStream(ctx, jetstream.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Storage:   jetstream.FileStorage,
		Retention: jetstream.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour,
		MaxBytes:  512 * 1024 * 1024,
	})
	return err
}
