# Laravel Go

A Go-based worker system compatible with Laravel's Queue and Schedule.

## Features

- **Queue Workers**: Consume jobs from Redis, Database, or SQS queues.
- **Job Handling**: Map Laravel job classes to Go handler functions.
- **PHP Serialization**: Support for `phpserialize` to read serialized PHP objects in job payloads.
- **Failed Jobs**: Automatically log failed jobs to a database table (compatible with Laravel's `failed_jobs`).
- **Scheduler**: A kernel scheduler similar to Laravel's, supporting `WithoutOverlapping` and `OnOneServer` using distributed locks.

## Installation

```bash
go get github.com/pixelvide/laravel-go
```

## Documentation & AI Context

For detailed package documentation, see [doc.go](doc.go) or run `go doc github.com/pixelvide/laravel-go`.

For AI agents or developers needing a quick overview of the codebase structure and import paths, refer to [AGENTS.md](AGENTS.md).

See [docs/register_jobs.md](docs/register_jobs.md) for details on registering job handlers.
See [docs/logging.md](docs/logging.md) for details on using the integrated logger and tracing.

## Usage

### 1. Define Handlers

Create a handler function that matches the `queue.Handler` signature:

```go
import "github.com/pixelvide/laravel-go/pkg/queue"

func MyHandler(ctx context.Context, job *queue.Job) error {
    log.Printf("Processing job: %s", job.Payload.DisplayName)
    return nil
}
```

### 2. Register Handlers

Register the handler with the Laravel job class name:

```go
queue.Register("App\\Jobs\\ProcessPodcast", MyHandler)
```

### 3. Start Worker

```go
package main

import (
    "context"
    "github.com/pixelvide/laravel-go/pkg/config"
    "github.com/pixelvide/laravel-go/pkg/driver/redis"
    "github.com/pixelvide/laravel-go/pkg/worker"
    "github.com/pixelvide/laravel-go/pkg/queue"
)

func main() {
    // Setup Redis Driver
    driver := redis.NewRedisDriver(config.RedisConfig{
        Addr: "localhost:6379",
    })

    // Setup Worker
    w := worker.NewWorker(driver, nil, "default", 5)

    // Run
    w.Run(context.Background())
}
```

### SQS Driver

To use Amazon SQS:

```go
import (
    "github.com/pixelvide/laravel-go/pkg/driver/sqs"
    "github.com/pixelvide/laravel-go/pkg/config"
)

// ...

sqsClient, _ := config.LoadSQSClient(ctx, config.SQSConfig{
    Region:   "us-east-1",
    QueueUrl: "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue",
})
driver := sqs.NewSQSDriver(sqsClient, "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue")
```

### Scheduler

The Scheduler allows running periodic tasks with distributed locking.

```go
import (
    "github.com/pixelvide/laravel-go/pkg/schedule"
    "github.com/pixelvide/laravel-go/pkg/driver/redis"
)

// ...

// Use Redis for distributed locks
redisClient := redis.NewRedisDriver(redisConfig).Client
lockProvider := schedule.NewRedisLockProvider(redisClient)

kernel := schedule.NewKernel(lockProvider)

// Register a task running every minute
kernel.Register("* * * * *", func() {
    log.Println("Running task...")
}, schedule.OnOneServer("unique-task-name"))

kernel.Run()
```
