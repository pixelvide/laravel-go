package console

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pixelvide/laravel-go/pkg/queue"
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

var (
	globalDriver         queue.Driver
	globalFailedProvider queue.FailedJobProvider
)

// SetDriver sets the queue driver for the worker command
func SetDriver(driver queue.Driver) {
	globalDriver = driver
}

// SetFailedJobProvider sets the failed job provider for the worker command
func SetFailedJobProvider(provider queue.FailedJobProvider) {
	globalFailedProvider = provider
}

var workerCmd = &cobra.Command{
	Use:     "queue:work",
	Aliases: []string{"worker"},
	Short:   "Start the queue worker",
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

		if globalDriver == nil {
			log.Fatal().Msg("No queue driver configured. Please call console.SetDriver() before executing the root command.")
		}

		// Initialize Worker
		w := worker.NewWorker(globalDriver, globalFailedProvider, queueName, concurrency, tracer)

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
