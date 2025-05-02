package rabbitmq

import "github.com/rabbitmq/amqp091-go"

type ConsumeOptions struct {
	Queue       string
	ConsumerTag string
	AutoAck     bool
	Exclusive   bool
	NoLocal     bool
	NoWait      bool
	Args        amqp091.Table
	Prefetch    int
}

func (r *RabbitMQ) Consume(opts ConsumeOptions) (<-chan amqp091.Delivery, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if opts.Prefetch > 0 {
		if err := r.channel.Qos(opts.Prefetch, 0, false); err != nil {
			return nil, err
		}
	}

	return r.channel.Consume(
		opts.Queue,
		opts.ConsumerTag,
		opts.AutoAck,
		opts.Exclusive,
		opts.NoLocal,
		opts.NoWait,
		opts.Args,
	)
}
