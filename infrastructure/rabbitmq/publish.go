package rabbitmq

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
)

type PublishOptions struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Headers    amqp091.Table
	Body       []byte
	Context    context.Context
}

func (r *RabbitMQ) Publish(opts PublishOptions) error {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.channel.PublishWithContext(
		ctx,
		opts.Exchange,
		opts.RoutingKey,
		opts.Mandatory,
		opts.Immediate,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         opts.Body,
			Headers:      opts.Headers,
			DeliveryMode: amqp091.Persistent,
		},
	)
}
