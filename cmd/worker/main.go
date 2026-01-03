package main

import (
	"context"

	"github.com/pixelvide/laravel-go/pkg/queue"
	"github.com/pixelvide/laravel-go/pkg/root"
	"github.com/pixelvide/laravel-go/pkg/telemetry"

	_ "github.com/pixelvide/laravel-go/pkg/console" // Register commands
)

// ExampleHandler is a sample job handler
func ExampleHandler(ctx context.Context, job *queue.Job) error {
	logger := telemetry.LoggerFromContext(ctx)
	logger.Info().
		Str("job_name", job.Payload.DisplayName).
		Str("uuid", job.Payload.UUID).
		Msg("Processing job")

	// Example of accessing unserialized PHP data
	if job.UnserializedData != nil {
		// Map properties if needed
		// props := queue.GetPHPProperty(job.UnserializedData, "podcastId")
		logger.Info().Msg("Unserialized data present")
	}
	return nil
}

func main() {
	// 1. Register Handlers
	// Register a handler for a hypothetical Laravel job "App\Jobs\ProcessPodcast"
	queue.Register("App\\Jobs\\ProcessPodcast", ExampleHandler)

	// 2. Execute Root Command
	root.Execute()
}
