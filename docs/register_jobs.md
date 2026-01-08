# Registering Jobs

To process Laravel jobs in Go, you need to register handlers that map to the Laravel job class names.

## Handler Signature

A job handler is a function with the following signature:

```go
func(ctx context.Context, job *queue.Job) error
```

*   `ctx`: The context, which includes tracing information and the logger.
*   `job`: The job object containing the payload and unserialized PHP data.

## Registering a Handler

You should register your handlers in your application's entry point (e.g., `main.go`) before executing the root command.

```go
package main

import (
    "context"
    "github.com/pixelvide/laravel-go/pkg/queue"
    "github.com/pixelvide/laravel-go/pkg/root"
    "github.com/pixelvide/laravel-go/pkg/telemetry"

    // Import console to register queue:work command
    _ "github.com/pixelvide/laravel-go/pkg/console"
)

func ProcessPodcast(ctx context.Context, job *queue.Job) error {
    // Get the logger
    logger := telemetry.LoggerFromContext(ctx)

    // Access job arguments (public properties of the Laravel job class)
    podcastID := job.GetArg("podcastId")

    logger.Info().Any("podcast_id", podcastID).Msg("Processing podcast")

    return nil
}

func main() {
    // Register the handler
    // The string must match the Laravel class name exactly
    queue.Register("App\\Jobs\\ProcessPodcast", ProcessPodcast)

    // Run the CLI
    root.Execute()
}
```

## Accessing Job Data

The `queue.Job` struct provides a helper method `GetArg(key string)` to access public properties of the unserialized PHP job object.

```go
if id := job.GetArg("id"); id != nil {
    // Use id
}
```

## Logging

The system uses `zerolog` and `OpenTelemetry`. You should retrieve the logger from the context to ensure logs are correlated with the trace ID and job ID.

```go
logger := telemetry.LoggerFromContext(ctx)
logger.Info().Msg("Log message")
```
