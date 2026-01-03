package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/driver/redis"
	"github.com/pixelvide/laravel-go/pkg/queue"
	"github.com/pixelvide/laravel-go/pkg/schedule"
	"github.com/pixelvide/laravel-go/pkg/telemetry"
	"github.com/pixelvide/laravel-go/pkg/worker"
	"github.com/rs/zerolog/log"
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
	// Parse command line flags
	queueName := flag.String("queue", "default", "Name of the queue to process")
	concurrency := flag.Int("workers", 5, "Number of concurrent workers")
	flag.Parse()

	// Initialize Telemetry
	telemetry.SetGlobalLogger()

	tp, err := telemetry.InitTracer("laravel-go-worker")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("Error shutting down tracer")
		}
	}()

	tracer := tp.Tracer("worker")

	// 1. Configure
	redisConfig := config.RedisConfig{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	// 2. Register Handlers
	// Register a handler for a hypothetical Laravel job "App\Jobs\\ProcessPodcast"
	queue.Register("App\\Jobs\\ProcessPodcast", ExampleHandler)

	// 3. Initialize Driver
	driver := redis.NewRedisDriver(redisConfig)

	// 4. Initialize Worker
	// For this example, we aren't setting up a database failed job provider, so we pass nil
	w := worker.NewWorker(driver, nil, *queueName, *concurrency, tracer)

	// 5. Run Worker with Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info().Msg("Shutting down worker...")
		cancel()
	}()

	log.Info().Str("queue", *queueName).Int("workers", *concurrency).Msg("Starting worker pool...")

	// Example: Run Scheduler (Optional)
	// go runScheduler(redisConfig)

	w.Run(ctx)
	log.Info().Msg("Worker pool stopped.")
}

// runScheduler is an example function for setting up the scheduler.
// It is unused in the default worker configuration but provided as a reference.
//
//nolint:unused
func runScheduler(redisCfg config.RedisConfig) {
	// Example of starting the scheduler
	redisClient := redis.NewRedisDriver(redisCfg).Client

	lockProvider := schedule.NewRedisLockProvider(redisClient)
	kernel := schedule.NewKernel(lockProvider)

	kernel.Register("* * * * *", func() {
		log.Info().Msg("Running scheduled task...")
	}, schedule.OnOneServer("my-scheduled-task"))

	kernel.Run()
}
