# AI Agent Context

This repository (`github.com/pixelvide/laravel-go`) provides a Go implementation of Laravel's Queue Worker and Task Scheduler. It is designed to be interoperable with Laravel applications.

## Import Paths

When working with this repository, use the following import paths:

*   **Core Interfaces & Registry**: `github.com/pixelvide/laravel-go/pkg/queue`
*   **Worker Pool**: `github.com/pixelvide/laravel-go/pkg/worker`
*   **Configuration**: `github.com/pixelvide/laravel-go/pkg/config`
*   **Drivers**:
    *   **Redis**: `github.com/pixelvide/laravel-go/pkg/driver/redis`
    *   **Database (SQL)**: `github.com/pixelvide/laravel-go/pkg/driver/database`
    *   **SQS**: `github.com/pixelvide/laravel-go/pkg/driver/sqs`
*   **Scheduler**: `github.com/pixelvide/laravel-go/pkg/schedule`

## Core Concepts

### 1. Queue Jobs
Jobs are consumed from a driver (`Pop`). The payload matches Laravel's JSON structure.
*   **Handler Signature**: `func(ctx context.Context, job *queue.Job) error`
*   **Registration**: Use `queue.Register("App\\Jobs\\ClassName", handlerFunc)` to map a Laravel class name to a Go function.
*   **Payload Access**:
    *   `job.Payload`: Parsed JSON envelope (DisplayName, UUID, Attempts, etc.).
    *   `job.UnserializedData`: If the job contains a serialized PHP command object (common in Laravel), this field contains the unserialized data (map or object).

### 2. Workers
The worker pool (`pkg/worker`) polls the driver.
*   It handles concurrency.
*   It respects job timeouts specified in the payload.
*   It handles retries (`attempts` < `maxTries`).
*   It logs failed jobs to a `FailedJobProvider` (usually Database) if `maxTries` is exceeded.

### 3. Drivers
Drivers implement the `queue.Driver` interface:
*   `Pop(ctx, queueName)`: Retrieve a job.
*   `Push(ctx, queueName, body)`: Dispatch a job.
*   `Ack(ctx, job)`: Acknowledge completion (delete from queue).

### 4. Scheduler
The `pkg/schedule` package mimics `App\Console\Kernel`.
*   **Kernel**: Manages cron jobs.
*   **LockProvider**: Interface for distributed locks (e.g., `RedisLockProvider`).
*   **Features**:
    *   `WithoutOverlapping()`: Prevents concurrent execution on the same machine.
    *   `OnOneServer(name)`: Uses distributed lock to run on only one server in the cluster.

## Directory Structure

```
.
├── cmd/
│   └── worker/         # Main entry point example
├── pkg/
│   ├── config/         # Configuration structs
│   ├── driver/         # Queue implementations (redis, database, sqs)
│   ├── queue/          # Core interfaces, registry, serializer
│   ├── schedule/       # Cron scheduler & distributed locking
│   └── worker/         # Worker pool logic
├── AGENTS.md           # This file
├── doc.go              # Go package documentation
└── go.mod
```

## coding Standards

*   Use `context.Context` for cancellation and timeouts.
*   Use structured logging where possible.
*   Ensure thread safety in drivers and registry.
