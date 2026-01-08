# Logging and Tracing

The `laravel-go` library provides integrated structured logging (using `zerolog`) and distributed tracing (using `OpenTelemetry`).

## Trace IDs and Span IDs

When a job is processed, the worker automatically:
1.  Starts a new OpenTelemetry trace.
2.  Creates a structured logger attached to the context.
3.  Injects the `trace_id` and `job_id` into the logger.

This ensures that every log message generated during the job execution is automatically tagged with the Trace ID, allowing you to correlate logs across different services or even within a single job execution.

## Using the Logger

To use the logger in your job handlers, you should retrieve it from the context using `telemetry.LoggerFromContext(ctx)`.

### Example

```go
package main

import (
    "context"
    "github.com/pixelvide/laravel-go/pkg/queue"
    "github.com/pixelvide/laravel-go/pkg/telemetry"
)

func ProcessOrder(ctx context.Context, job *queue.Job) error {
    // 1. Get the logger from the context
    // This logger already has "trace_id" and "job_uuid" fields set.
    logger := telemetry.LoggerFromContext(ctx)

    orderID := job.GetArg("orderId")

    // 2. Log messages
    // These logs will include the trace context automatically.
    logger.Info().
        Any("order_id", orderID).
        Msg("Starting to process order")

    if err := processOrder(orderID); err != nil {
        logger.Error().Err(err).Msg("Failed to process order")
        return err
    }

    logger.Info().Msg("Order processed successfully")
    return nil
}
```

### Log Output Example

The output will look something like this (formatted for readability):

```json
{
  "level": "info",
  "time": "2023-10-27T10:00:00Z",
  "service": "LaravelGoApp",
  "command": "queue:work",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "job_uuid": "550e8400-e29b-41d4-a716-446655440000",
  "job_name": "App\\Jobs\\ProcessOrder",
  "order_id": 12345,
  "message": "Starting to process order"
}
```

## Configuring Telemetry

The telemetry system is initialized automatically when using the `queue:work` command. You can customize the behavior by setting the global logger or tracer provider in your application setup if needed, but the default setup is sufficient for most use cases.
