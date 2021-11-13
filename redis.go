package lock

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisLocker struct {
	key    string
	client *redis.Client
}

func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{
		key:    lockKey,
		client: client,
	}
}

func (l *RedisLocker) GetLock(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	result, err := l.client.SetNX(ctx, l.key, locked, -1).Result()
	if err != nil {
		cancel()
		return err
	}
	if !result {
		cancel()
		return errors.New("failed to lock")
	}
	return nil
}

func (l *RedisLocker) HasLock(ctx context.Context) (bool, error) {
	_, err := l.client.Get(ctx, l.key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (l *RedisLocker) ReleaseLock(ctx context.Context) error {
	_, err := l.client.Del(ctx, l.key).Result()
	return err
}
