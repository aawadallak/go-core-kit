package rmq

import (
	"encoding/json"
	"strings"

	"context"

	"github.com/aawadallak/go-core-kit/core/logger"
	"github.com/wagslane/go-rabbitmq"
)

type Publisher struct {
	conn   *rabbitmq.Conn
	closed bool
}

type PublisherConfig struct {
	Endpoint string
}

func (p *Publisher) Close() error {
	if p.closed {
		return nil
	}

	if err := p.conn.Close(); err != nil {
		return err
	}

	if p.conn != nil {
		p.conn.Close() //nolint:gosec // G104: best-effort cleanup on shutdown
	}

	p.closed = true

	return nil
}

type PublishMessage struct {
	Message any
	Queue   string
}

func (p *Publisher) Publish(ctx context.Context, req PublishMessage) error {
	publisher, err := rabbitmq.NewPublisher(
		p.conn,
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		return err
	}

	defer publisher.Close()

	message, err := json.Marshal(req.Message)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debugf("publish message to queue %s", strings.ToLower(req.Queue)+"-exchange")

	return publisher.Publish(
		message,
		[]string{""},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(strings.ToLower(req.Queue)+"-exchange"),
		rabbitmq.WithPublishOptionsPersistentDelivery,
	)
}

func NewPublisher(cfg PublisherConfig) (*Publisher, error) {
	conn, err := rabbitmq.NewConn(
		cfg.Endpoint,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:   conn,
		closed: false,
	}, nil
}
