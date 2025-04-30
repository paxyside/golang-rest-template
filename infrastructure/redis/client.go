package redis

import (
	"context"
	"emperror.dev/errors"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"sync"
	"time"
)

type Storage interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Delete(key string) error
}

type Redis struct {
	client *redis.Client
	mu     sync.Mutex
}

func Init(redisURI string) (*Redis, error) {
	opt, err := redis.ParseURL(redisURI)
	if err != nil {
		return nil, errors.Wrap(err, "redis.ParseURL")
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, errors.Wrap(err, "client.Ping")
	}

	r := &Redis{
		client: client,
	}

	go r.connectionWorker(redisURI)

	return r, nil
}

func (r *Redis) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.client.Close(); err != nil {
		slog.Error("failed to close redis connection", slog.Any("error", err))
	}
}

func (r *Redis) connectionWorker(redisURI string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := r.client.Ping(ctx).Err()
		cancel()

		if err != nil {
			slog.Error("redis ping failed", slog.Any("error", err))

			r.mu.Lock()
			newOpt, _ := redis.ParseURL(redisURI)
			newClient := redis.NewClient(newOpt)

			if err := newClient.Ping(context.Background()).Err(); err == nil {
				_ = r.client.Close()
				r.client = newClient
				slog.Info("redis reconnected")
			} else {
				slog.Error("redis reconnect failed", slog.Any("error", err))
			}

			r.mu.Unlock()
		}
	}
}

func (r *Redis) Set(key string, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return errors.Wrap(err, "r.Client.Set")
	}

	return nil
}

func (r *Redis) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", errors.Wrap(err, "r.Client.Get")
	}

	return val, nil
}

func (r *Redis) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return errors.Wrap(err, "r.Client.Del")
	}

	return nil
}
