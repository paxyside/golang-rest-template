package rabbitmq

import (
	"golang-template/infrastructure/rabbitmq"

	"github.com/rabbitmq/amqp091-go"
)

type Queue interface {
	DeclareQueue(cfg rabbitmq.QueueConfig) (amqp091.Queue, error)
	Publish(opts rabbitmq.PublishOptions) error
	Consume(opts rabbitmq.ConsumeOptions) (<-chan amqp091.Delivery, error)
	Close()
}
