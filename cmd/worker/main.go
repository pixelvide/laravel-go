package main

import (
	"context"
	"fmt"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/queue"
	"github.com/pixelvide/laravel-go/pkg/root"
	"github.com/pixelvide/laravel-go/pkg/schedule"
	"github.com/pixelvide/laravel-go/pkg/telemetry"
	"github.com/rs/zerolog/log"
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
	// 1. Load Configuration (Optional, for app usage)
	// You can load the config here to use it in your own code.
	// The queue:work command loads it automatically, so this is just for demonstration.
	cfg, err := config.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Could not load config")
	} else {
		log.Info().Str("app_name", cfg.App.Name).Str("env", cfg.App.Env).Msg("Loaded application config")
	}

	// 2. Register Handlers
	// Register a handler for a hypothetical Laravel job "App\Jobs\ProcessPodcast"
	queue.Register("App\\Jobs\\ProcessPodcast", ExampleHandler)

	// 3. Register Scheduled Tasks
	// Example: Run every minute on one server
	schedule.Register("* * * * *", func() {
		fmt.Println("Running scheduled task: Every Minute")
	}, schedule.OnOneServer("every-minute-task"))

	// 4. Register Custom Commands
	// Example: A custom "hello" command
	helloCmd := &cobra.Command{
		Use:   "hello",
		Short: "Prints a hello message",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Hello from %s!\n", cfg.App.Name)
		},
	}
	root.GetRoot().AddCommand(helloCmd)

	// 5. Execute Root Command
	root.Execute()
}
