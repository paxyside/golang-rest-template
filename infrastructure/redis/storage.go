package redis

import (
	"context"
	"time"

	"emperror.dev/errors"
)

func (r *Redis) Set(
	ctx context.Context, key string, value string, ttl time.Duration,
) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return errors.Wrap(err, "r.Client.Set")
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", errors.Wrap(err, "r.Client.Get")
	}

	return val, nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return errors.Wrap(err, "r.Client.Del")
	}

	return nil
}
