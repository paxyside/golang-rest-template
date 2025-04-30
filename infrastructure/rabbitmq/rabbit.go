package rabbitmq

import (
	"context"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
	"sync"
	"time"
)

type Queue interface {
	DeclareQueue(name string) (amqp091.Queue, error)
	Publish(ctx context.Context, queue string, body []byte) error
	Consume(queue string) (<-chan amqp091.Delivery, error)
	Close()
}

type RabbitMQ struct {
	uri     string
	conn    *amqp091.Connection
	channel *amqp091.Channel
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
}

func Init(uri string) (*RabbitMQ, error) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &RabbitMQ{
		uri:    uri,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := r.connect(); err != nil {
		cancel()
		return nil, err
	}

	go r.monitorConnection()
	return r, nil
}

func (r *RabbitMQ) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conn, err := amqp091.Dial(r.uri)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.conn = conn
	r.channel = channel
	return nil
}

func (r *RabbitMQ) monitorConnection() {
	notifyClose := r.conn.NotifyClose(make(chan *amqp091.Error, 1))

	for {
		select {
		case err := <-notifyClose:
			if err != nil {
				slog.Error("RabbitMQ connection closed", slog.Any("error", err))
			}
			slog.Info("attempting to reconnect to RabbitMQ")

			for {
				select {
				case <-r.ctx.Done():
					return
				default:
					if err := r.connect(); err != nil {
						slog.Error("reconnect failed", slog.Any("error", err))
						time.Sleep(5 * time.Second)
						continue
					}
					slog.Info("successfully reconnected to RabbitMQ")
					notifyClose = r.conn.NotifyClose(make(chan *amqp091.Error, 1))
					break
				}
			}
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *RabbitMQ) DeclareQueue(name string) (amqp091.Queue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) Publish(ctx context.Context, queue string, body []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.channel.PublishWithContext(ctx,
		"",
		queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (r *RabbitMQ) Consume(queue string) (<-chan amqp091.Delivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.channel.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) Close() {
	r.cancel()
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
