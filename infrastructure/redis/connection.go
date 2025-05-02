package redis

import (
	"context"
	"golang-template/internal/domain/logger"
	"log/slog"
	"sync"
	"time"

	"github.com/spf13/viper"

	"emperror.dev/errors"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	mu sync.RWMutex

	uri           string
	client        *redis.Client
	l             logger.Loggerer
	checkInterval time.Duration
	connTimeout   time.Duration
}

func Init(ctx context.Context, redisURI string, l logger.Loggerer) (*Redis, error) {
	r := &Redis{
		uri:           redisURI,
		l:             l,
		checkInterval: viper.GetDuration("app.redis.healthcheck_interval"),
		connTimeout:   viper.GetDuration("app.redis.connection_timeout"),
	}

	if err := r.connect(ctx); err != nil {
		return nil, errors.Wrap(err, "r.connect")
	}

	go r.healthChecker(ctx)

	return r, nil
}

func (r *Redis) healthChecker(ctx context.Context) {
	ticker := time.NewTicker(r.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.Ping(ctx); err != nil {
				r.l.Error("redis ping failed", slog.Any("error", err))

				if err = r.connect(ctx); err != nil {
					r.l.Error("redis reconnect failed", slog.Any("error", err))
				}
			}
		}
	}
}

func (r *Redis) connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	newOpt, err := redis.ParseURL(r.uri)
	if err != nil {
		return errors.Wrap(err, "redis.ParseURL")
	}

	newClient := redis.NewClient(newOpt)

	if err = newClient.Ping(ctx).Err(); err != nil {
		return errors.Wrap(err, "newClient.Ping")
	}

	if r.client != nil {
		_ = r.client.Close()
	}

	r.client = newClient

	return nil
}

func (r *Redis) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.connTimeout)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return errors.Wrap(err, "r.Client.Ping")
	}

	return nil
}

func (r *Redis) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.client == nil {
		return
	}

	err := r.client.Close()
	if err != nil {
		r.l.Error("failed to close redis connection", slog.Any("error", err))
	} else {
		r.l.Info("redis connection closed")
	}

	r.client = nil
}
