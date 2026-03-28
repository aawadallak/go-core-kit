// Package rmq provides rmq functionality.
package rmq

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/wagslane/go-rabbitmq"
)

type Config struct {
	Endpoint  string
	QueueName string
}

type Consumer struct {
	config     Config
	conn       *rabbitmq.Conn
	consumer   *rabbitmq.Consumer
	closed     bool
	init       bool
	stream     chan string
	commit     chan bool
	processing bool
}

func (c *Consumer) Close() error {
	if c.closed {
		return nil
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	if c.consumer != nil {
		c.consumer.Close()
	}

	c.closed = true
	close(c.stream)

	return nil
}

func (c *Consumer) Subscribe(ctx context.Context) (string, error) {
	if c.closed {
		return "", fmt.Errorf("subscriber is closed: %v", c.closed)
	}

	if c.processing {
		c.commit <- false
	}

	if c.init {
		return <-c.stream, nil
	}

	handler := func(d rabbitmq.Delivery) rabbitmq.Action {
		c.stream <- string(d.Body)
		c.processing = true

		shouldCommit := <-c.commit
		if !shouldCommit {
			return rabbitmq.NackRequeue
		}

		return rabbitmq.Ack
	}

	log.Println("starting consumer on queue: ", strings.ToLower(c.config.QueueName))

	options := []func(*rabbitmq.ConsumerOptions){
		rabbitmq.WithConsumerOptionsExchangeName(strings.ToLower(c.config.QueueName) + "-exchange"),
		rabbitmq.WithConsumerOptionsQueueDurable,
	}

	consumer, err := rabbitmq.NewConsumer(c.conn, c.config.QueueName, options...)
	if err != nil {
		return "", err
	}

	if err := consumer.Run(handler); err != nil {
		return "", err
	}

	c.consumer = consumer
	c.init = true

	return <-c.stream, nil
}

func (c *Consumer) Commit(ctx context.Context, message any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.commit <- true
	c.processing = false

	return nil
}

func NewConsumer(cfg Config) (*Consumer, error) {
	conn, err := rabbitmq.NewConn(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		config: cfg,
		conn:   conn,
		stream: make(chan string, 1),
		commit: make(chan bool, 1),
	}, nil
}
