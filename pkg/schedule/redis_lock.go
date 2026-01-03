package schedule

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLockProvider implements LockProvider using Redis SETNX
type RedisLockProvider struct {
	client *redis.Client
}

func NewRedisLockProvider(client *redis.Client) *RedisLockProvider {
	return &RedisLockProvider{client: client}
}

func (r *RedisLockProvider) GetLock(ctx context.Context, name string, duration time.Duration) (bool, error) {
	// SET name value NX EX duration
	success, err := r.client.SetNX(ctx, "schedule_lock:"+name, "locked", duration).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

func (r *RedisLockProvider) ReleaseLock(ctx context.Context, name string) error {
	return r.client.Del(ctx, "schedule_lock:"+name).Err()
}
