package redis

import (
	"context"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/queue"
	goredis "github.com/redis/go-redis/v9"
)

type RedisDriver struct {
	client *goredis.Client
}

// NewRedisDriver creates a new Redis driver instance
func NewRedisDriver(cfg config.RedisConfig) *RedisDriver {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &RedisDriver{client: rdb}
}

// Pop blocks until a job is available and returns it
func (r *RedisDriver) Pop(ctx context.Context, queueName string) (*queue.Job, error) {
	// BLPOP returns [key, value]
	// We use a timeout of 0 to block indefinitely until a job is available,
	// but we respect the context deadline if set.

	// Since BLPOP with 0 timeout blocks forever, we might want to use a small timeout loop
	// to allow context cancellation checks, or rely on the client respecting context.
	// go-redis BLPOP respects context.

	// Laravel queue names in redis usually have a prefix.
	// If the user passes "default", the key is likely "queues:default".
	// But we will assume the config passes the full key or we handle it here.
	// Standard Laravel uses "queues:{queueName}".
	// Let's assume the user configures the full key or handles the prefix logic outside.
	// For now, we use queueName as is.

	result, err := r.client.BLPop(ctx, 0, queueName).Result()
	if err != nil {
		return nil, err
	}

	// result[0] is the key (queueName), result[1] is the value (payload)
	if len(result) < 2 {
		return nil, context.DeadlineExceeded // Should not happen with successful BLPop
	}

	return &queue.Job{
		ID:   "", // Redis lists don't have explicit IDs unless inside the body
		Body: []byte(result[1]),
	}, nil
}

// Push adds a job to the queue
func (r *RedisDriver) Push(ctx context.Context, queueName string, body []byte) error {
	return r.client.RPush(ctx, queueName, body).Err()
}

// Fail pushes the job to a failed jobs list.
// In a real Laravel setup, this is usually a database table (failed_jobs).
// Since we are using Redis here, we will push to a "failed" list.
func (r *RedisDriver) Fail(ctx context.Context, queueName string, body []byte, err error) error {
	// TODO: Wrap body in a failed job structure with exception details?
	// For now, simply move to a failed list.
	failedQueue := queueName + ":failed"
	return r.client.RPush(ctx, failedQueue, body).Err()
}
