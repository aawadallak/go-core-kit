package natsjetstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type WorkerConfig[T any] struct {
	Endpoint      string
	StreamName    string
	Subject       string
	DurableName   string
	DLQStreamName string
	DLQSubject    string
	MaxDeliver    int
	AckWait       time.Duration
	FetchMaxWait  time.Duration
	Handler       func(ctx context.Context, message *T) error
}

type Worker[T any] struct {
	conn     *nats.Conn
	js       jetstream.JetStream
	consumer jetstream.Consumer
	cfg      WorkerConfig[T]
	closed   atomic.Bool
	closeCh  chan struct{}
	wg       sync.WaitGroup
	closeErr error
	closeMux sync.Mutex
	closeOne sync.Once
}

func (w *Worker[T]) hasDLQ() bool {
	return w.cfg.DLQStreamName != "" && w.cfg.DLQSubject != ""
}

func NewWorker[T any](ctx context.Context, cfg WorkerConfig[T]) (*Worker[T], error) {
	conn, err := nats.Connect(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if cfg.MaxDeliver <= 0 {
		cfg.MaxDeliver = 5
	}
	if cfg.AckWait <= 0 {
		cfg.AckWait = 30 * time.Second
	}
	if cfg.FetchMaxWait <= 0 {
		cfg.FetchMaxWait = 1 * time.Second
	}

	if err := ensureStream(ctx, js, cfg.StreamName, []string{cfg.Subject}); err != nil {
		conn.Close()
		return nil, err
	}
	if cfg.DLQStreamName != "" && cfg.DLQSubject != "" {
		if err := ensureStream(ctx, js, cfg.DLQStreamName, []string{cfg.DLQSubject}); err != nil {
			conn.Close()
			return nil, err
		}
	}

	stream, err := js.Stream(ctx, cfg.StreamName)
	if err != nil {
		conn.Close()
		return nil, err
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       cfg.DurableName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       cfg.AckWait,
		MaxDeliver:    cfg.MaxDeliver,
		FilterSubject: cfg.Subject,
	})
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Worker[T]{
		conn:     conn,
		js:       js,
		consumer: consumer,
		cfg:      cfg,
		closeCh:  make(chan struct{}),
	}, nil
}

func (w *Worker[T]) Start(ctx context.Context) {
	for {
		if w.shouldStop(ctx) {
			return
		}

		msgs, err := w.consumer.Fetch(1, jetstream.FetchMaxWait(w.cfg.FetchMaxWait))
		if err != nil {
			if w.shouldStop(ctx) {
				return
			}
			continue
		}

		for msg := range msgs.Messages() {
			w.wg.Add(1)
			func(message jetstream.Msg) {
				defer w.wg.Done()

				if err := w.handleMessage(ctx, message); err != nil {
					log.Println(err)
				}
			}(msg)
		}
	}
}

func (w *Worker[T]) shouldStop(ctx context.Context) bool {
	if ctx.Err() != nil {
		return true
	}

	if w.closed.Load() {
		return true
	}

	select {
	case <-w.closeCh:
		return true
	default:
		return false
	}
}

func (w *Worker[T]) handleMessage(ctx context.Context, msg jetstream.Msg) error {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicErr := fmt.Errorf("panic while handling message: %v", recovered)
			log.Println(panicErr)

			if w.hasDLQ() {
				if err := publishDLQ(ctx, w.js, w.cfg.DLQSubject, msg.Data(), panicErr.Error()); err != nil {
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

	var payload T
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		if w.hasDLQ() {
			if dlqErr := publishDLQ(ctx, w.js, w.cfg.DLQSubject, msg.Data(), "invalid_json"); dlqErr != nil {
				log.Printf("failed to publish invalid_json to DLQ: %v", dlqErr)
			}
		}
		return msg.Ack()
	}

	if err := w.cfg.Handler(ctx, &payload); err != nil {
		if w.hasDLQ() {
			if dlqErr := publishDLQ(ctx, w.js, w.cfg.DLQSubject, msg.Data(), err.Error()); dlqErr != nil {
				log.Printf("failed to publish handler error to DLQ: %v", dlqErr)
			}
			return msg.Ack()
		}
		return msg.Nak()
	}

	return msg.Ack()
}

func (w *Worker[T]) Close() error {
	w.closeOne.Do(func() {
		w.closed.Store(true)
		close(w.closeCh)

		// Wait for in-flight handlers (and ack/nack) to complete.
		w.wg.Wait()

		if w.conn == nil {
			return
		}

		if err := w.conn.Drain(); err != nil {
			w.conn.Close()

			// Closing a drained/closed connection can return benign errors.
			if !errors.Is(err, nats.ErrConnectionClosed) && !errors.Is(err, nats.ErrConnectionDraining) {
				w.closeMux.Lock()
				w.closeErr = err
				w.closeMux.Unlock()
			}
		}
	})

	w.closeMux.Lock()
	defer w.closeMux.Unlock()
	return w.closeErr
}

func publishDLQ(ctx context.Context, js jetstream.JetStream, subject string, payload []byte, reason string) error {
	envelope := map[string]any{
		"reason":      reason,
		"payload":     json.RawMessage(payload),
		"failed_at":   time.Now().UTC().Format(time.RFC3339Nano),
		"broker":      "nats-jetstream",
		"dlq_subject": subject,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	_, err = js.Publish(ctx, subject, data)
	return err
}
