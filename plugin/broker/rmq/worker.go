package rmq

import (
	"context"
	"encoding/json"
	"log"
)

type Worker[T any] struct {
	consumer *Consumer
	fn       func(ctx context.Context, message *T) error
	closed   bool
}

type WorkerConfig[T any] struct {
	Endpoint  string
	QueueName string
	Handler   func(ctx context.Context, message *T) error
}

func NewWorker[T any](config WorkerConfig[T]) (*Worker[T], error) {
	rmqConfig := Config{
		Endpoint:  config.Endpoint,
		QueueName: config.QueueName,
	}

	consumer, err := NewConsumer(rmqConfig)
	if err != nil {
		return nil, err
	}

	return &Worker[T]{
		consumer: consumer,
		fn:       config.Handler,
	}, nil
}

func (b *Worker[T]) Start(ctx context.Context) {
	for !b.closed {
		message, err := b.consumer.Subscribe(ctx)
		if err != nil {
			log.Println(err)
			continue
		}

		var req T
		if err := json.Unmarshal([]byte(message), &req); err != nil {
			log.Println(err)
			continue
		}

		if err := b.fn(ctx, &req); err != nil {
			log.Println(err)
			continue
		}

		if err := b.consumer.Commit(ctx, message); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (b *Worker[T]) Close() error {
	b.closed = true
	return b.consumer.Close()
}
