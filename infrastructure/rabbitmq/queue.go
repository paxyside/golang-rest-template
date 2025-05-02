package rabbitmq

import "github.com/rabbitmq/amqp091-go"

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp091.Table
}

func (r *RabbitMQ) DeclareQueue(cfg QueueConfig) (amqp091.Queue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.channel.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		cfg.Args,
	)
}
