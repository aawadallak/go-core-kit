package nats

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
	Endpoint      string
	Subject       string
	ConsumerGroup string
	Handler       func(ctx context.Context, message *T) error
}

func NewWorker[T any](config WorkerConfig[T]) (*Worker[T], error) {
	consumer, err := NewConsumer(Config{
		Endpoint:      config.Endpoint,
		Subject:       config.Subject,
		ConsumerGroup: config.ConsumerGroup,
	})
	if err != nil {
		return nil, err
	}

	return &Worker[T]{
		consumer: consumer,
		fn:       config.Handler,
	}, nil
}

func (w *Worker[T]) Start(ctx context.Context) {
	for {
		if w.closed {
			return
		}

		message, err := w.consumer.Subscribe(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Println(err)
			continue
		}

		var req T
		if err := json.Unmarshal([]byte(message), &req); err != nil {
			log.Println(err)
			continue
		}

		if err := w.fn(ctx, &req); err != nil {
			log.Println(err)
			continue
		}

		if err := w.consumer.Commit(ctx, message); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (w *Worker[T]) Close() error {
	w.closed = true
	return w.consumer.Close()
}
