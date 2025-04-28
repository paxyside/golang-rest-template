package rabbit

import (
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
	"sync"
)

type RabbitMQ struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	mu      sync.Mutex
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		if err := conn.Close(); err != nil {
			return nil, fmt.Errorf("failed to close connection: %w", err)
		}

		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
	}, nil
}

func (r *RabbitMQ) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			slog.Error("failed to close channel", slog.Any("error", err))
		}
	}

	if r.conn != nil {
		if err := r.channel.Close(); err != nil {
			slog.Error("failed to close channel", slog.Any("error", err))
		}
	}
}

func (r *RabbitMQ) DeclareQueue(name string) (amqp091.Queue, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	queue, err := r.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return amqp091.Queue{}, fmt.Errorf("failed to declare queue: %w", err)
	}

	return queue, nil
}

func (r *RabbitMQ) Publish(queue string, body []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.channel.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQ) Consume(queue string) (<-chan amqp091.Delivery, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	msgs, err := r.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	return msgs, nil
}
