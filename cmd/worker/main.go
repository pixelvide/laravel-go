package main

import (
	"context"
	"fmt"

	"github.com/pixelvide/laravel-go/pkg/queue"
	"github.com/pixelvide/laravel-go/pkg/root"
	"github.com/pixelvide/laravel-go/pkg/telemetry"
	"github.com/spf13/cobra"

	// Ensure drivers are loaded
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	// Ensure console commands are registered
	_ "github.com/pixelvide/laravel-go/pkg/console"
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
	// 1. Setup
	// We rely on .env configuration for the driver now.
	// You can still manually call console.SetDriver() if needed,
	// but by default it will auto-configure based on environment.

	// 2. Register Handlers
	// Register a handler for a hypothetical Laravel job "App\Jobs\ProcessPodcast"
	queue.Register("App\\Jobs\\ProcessPodcast", ExampleHandler)

	// 3. Register Custom Commands
	// Example: A custom "hello" command
	helloCmd := &cobra.Command{
		Use:   "hello",
		Short: "Prints a hello message",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from custom command!")
		},
	}
	root.GetRoot().AddCommand(helloCmd)

	// 4. Execute Root Command
	root.Execute()
}
