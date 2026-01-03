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
	// If the Laravel job class has a public property $podcastId
	if podcastID := job.GetArg("podcastId"); podcastID != nil {
		logger.Info().Any("podcast_id", podcastID).Msg("Found podcast ID in job arguments")
	} else {
		// Just to show data is present
		if job.UnserializedData != nil {
			logger.Info().Msg("Unserialized data present but podcastId not found")
		}
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
