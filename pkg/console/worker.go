package console

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/driver/redis"
	"github.com/pixelvide/laravel-go/pkg/root"
	"github.com/pixelvide/laravel-go/pkg/telemetry"
	"github.com/pixelvide/laravel-go/pkg/worker"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	queueName   string
	concurrency int
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the queue worker",
	Run: func(cmd *cobra.Command, args []string) {
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

		// Configure
		redisConfig := config.RedisConfig{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}

		// Initialize Driver
		driver := redis.NewRedisDriver(redisConfig)

		// Initialize Worker
		w := worker.NewWorker(driver, nil, queueName, concurrency, tracer)

		// Run Worker with Graceful Shutdown
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

		log.Info().Str("queue", queueName).Int("workers", concurrency).Msg("Starting worker pool...")

		w.Run(ctx)
		log.Info().Msg("Worker pool stopped.")
	},
}

func init() {
	workerCmd.Flags().StringVar(&queueName, "queue", "default", "Name of the queue to process")
	workerCmd.Flags().IntVar(&concurrency, "workers", 5, "Number of concurrent workers")

	root.GetRoot().AddCommand(workerCmd)
}
