package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/console"
	"github.com/pixelvide/laravel-go/pkg/driver/redis"
	"github.com/pixelvide/laravel-go/pkg/queue"
	"github.com/pixelvide/laravel-go/pkg/root"
	"github.com/pixelvide/laravel-go/pkg/telemetry"
	"github.com/spf13/cobra"
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
	// 1. Setup Driver
	// Configure Redis from environment or default
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	redisConfig := config.RedisConfig{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	}
	driver := redis.NewRedisDriver(redisConfig)

	// Set the driver for the worker command
	console.SetDriver(driver)

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
