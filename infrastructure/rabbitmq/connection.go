package rabbitmq

import (
	"context"
	"golang-template/internal/domain/logger"
	"log/slog"
	"sync"
	"time"

	"emperror.dev/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

type RabbitMQ struct {
	mu sync.RWMutex

	uri                 string
	conn                *amqp091.Connection
	channel             *amqp091.Channel
	ctx                 context.Context
	cancel              context.CancelFunc
	l                   logger.Loggerer
	connTimeout         time.Duration
	healthcheckInterval time.Duration
}

func Init(ctx context.Context, uri string, l logger.Loggerer) (*RabbitMQ, error) {
	rCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	r := &RabbitMQ{
		uri:                 uri,
		ctx:                 rCtx,
		cancel:              cancel,
		l:                   l,
		connTimeout:         viper.GetDuration("app.rabbitmq.connection_timeout"),
		healthcheckInterval: viper.GetDuration("app.rabbitmq.healthcheck_interval"),
	}

	if err := r.connect(); err != nil {
		return nil, errors.Wrap(err, "initial connect")
	}

	go r.healthChecker()

	return r, nil
}

func (r *RabbitMQ) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		_ = r.channel.Close()
		r.channel = nil
	}
	if r.conn != nil {
		_ = r.conn.Close()
		r.conn = nil
	}

	conn, err := amqp091.DialConfig(
		r.uri,
		amqp091.Config{
			Dial: amqp091.DefaultDial(r.connTimeout),
		},
	)
	if err != nil {
		return errors.Wrap(err, "amqp091.Dial")
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return errors.Wrap(err, "conn.Channel")
	}

	r.conn = conn
	r.channel = channel

	return nil
}

func (r *RabbitMQ) healthChecker() {
	for {
		r.mu.RLock()
		notifyClose := r.conn.NotifyClose(make(chan *amqp091.Error, 1))
		r.mu.RUnlock()

		select {
		case <-r.ctx.Done():
			return
		case err := <-notifyClose:
			if err != nil {
				r.l.Error("RabbitMQ connection closed", slog.Any("error", err))
			}

			r.l.Info("attempting to reconnect to RabbitMQ")

		reconnectLoop:
			for {
				select {
				case <-r.ctx.Done():
					return
				default:
					if err := r.connect(); err != nil {
						r.l.Error("reconnect failed", slog.Any("error", err))
						time.Sleep(r.healthcheckInterval)
						continue
					}

					r.l.Info("successfully reconnected to RabbitMQ")
					break reconnectLoop
				}
			}
		}
	}
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
