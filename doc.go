// Package laravelgo provides a Go-based worker system compatible with Laravel's Queue and Schedule.
//
// It allows consuming jobs from Redis, Database, or SQS queues, mapping Laravel job classes to Go handler functions,
// and scheduling tasks with distributed locking.
//
// Key subpackages:
//
//	github.com/pixelvide/laravel-go/pkg/queue     - Core interfaces, job registry, and payload handling
//	github.com/pixelvide/laravel-go/pkg/worker    - Worker pool implementation
//	github.com/pixelvide/laravel-go/pkg/schedule  - Kernel scheduler (Cron + Distributed Locks)
//	github.com/pixelvide/laravel-go/pkg/driver    - Queue drivers (redis, database, sqs)
//	github.com/pixelvide/laravel-go/pkg/config    - Configuration structs
//
// Example Usage:
//
//	package main
//
//	import (
//		"context"
//		"github.com/pixelvide/laravel-go/pkg/config"
//		"github.com/pixelvide/laravel-go/pkg/driver/redis"
//		"github.com/pixelvide/laravel-go/pkg/queue"
//		"github.com/pixelvide/laravel-go/pkg/worker"
//	)
//
//	func MyHandler(ctx context.Context, job *queue.Job) error {
//		// Process job...
//		return nil
//	}
//
//	func main() {
//		queue.Register("App\\Jobs\\MyJob", MyHandler)
//		driver := redis.NewRedisDriver(config.RedisConfig{Addr: "localhost:6379"})
//		w := worker.NewWorker(driver, nil, "default", 5)
//		w.Run(context.Background())
//	}
package laravelgo
