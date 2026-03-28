// Package nats provides nats functionality.
package nats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Config struct {
	Endpoint      string
	Subject       string
	ConsumerGroup string
}

type Consumer struct {
	config Config
	conn   *nats.Conn
	sub    *nats.Subscription
}

func NewConsumer(cfg Config) (*Consumer, error) {
	conn, err := nats.Connect(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		config: cfg,
		conn:   conn,
	}, nil
}

func (c *Consumer) ensureSubscription() error {
	if c.sub != nil {
		return nil
	}

	var (
		sub *nats.Subscription
		err error
	)

	if c.config.ConsumerGroup != "" {
		sub, err = c.conn.QueueSubscribeSync(c.config.Subject, c.config.ConsumerGroup)
	} else {
		sub, err = c.conn.SubscribeSync(c.config.Subject)
	}
	if err != nil {
		return err
	}

	c.sub = sub
	return nil
}

func (c *Consumer) Subscribe(ctx context.Context) (string, error) {
	if err := c.ensureSubscription(); err != nil {
		return "", err
	}

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		msg, err := c.sub.NextMsg(500 * time.Millisecond)
		if err == nil {
			return string(msg.Data), nil
		}
		if errors.Is(err, nats.ErrTimeout) {
			continue
		}
		return "", fmt.Errorf("nats subscribe: %w", err)
	}
}

func (c *Consumer) Commit(ctx context.Context, _ any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (c *Consumer) Close() error {
	if c.sub != nil {
		if err := c.sub.Unsubscribe(); err != nil {
			return err
		}
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
