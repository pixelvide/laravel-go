package console

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/database"
	driverdatabase "github.com/pixelvide/laravel-go/pkg/driver/database"
	"github.com/pixelvide/laravel-go/pkg/driver/redis"
	driversqs "github.com/pixelvide/laravel-go/pkg/driver/sqs"
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

		// Load Configuration
		cfg, err := config.Load()
		appName := "laravel-go"
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load configuration from .env")
		} else {
			appName = cfg.App.Name
			// Auto-configure Driver if not manually set
			if globalDriver == nil {
				d, err := configureDriver(cfg)
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to configure queue driver")
				}
				globalDriver = d
			}
		}

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
			log.Fatal().Msg("No queue driver configured. Please set QUEUE_CONNECTION in .env or call console.SetDriver().")
		}

		// Initialize Worker
		w := worker.NewWorker(globalDriver, globalFailedProvider, queueName, concurrency, appName, tracer)

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

func configureDriver(cfg *config.Config) (queue.Driver, error) {
	switch cfg.Queue.Connection {
	case "redis":
		rCfg := config.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}
		return redis.NewRedisDriver(rCfg), nil

	case "database":
		// Create DB Connection
		dbFactory := database.NewFactory()
		db, err := dbFactory.Connect(cfg.Database)
		if err != nil {
			return nil, err
		}
		return driverdatabase.NewDatabaseDriver(cfg.Database, db), nil

	case "sqs":
		ctx := context.Background()
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, err
		}
		client := awssqs.NewFromConfig(awsCfg)
		// We use the configured queue URL or name
		// For now, we assume QUEUE_QUEUE in env is the URL or we resolve it
		// This driver expects a URL.
		queueUrl := cfg.Queue.Queue
		if queueUrl == "default" {
			// Try to find URL for default queue?
			// Simplification: Assume user puts URL in QUEUE_QUEUE for SQS
			log.Warn().Msg("Using 'default' as SQS queue URL. Ensure QUEUE_QUEUE is set to a valid URL in .env")
		}
		return driversqs.NewSQSDriver(client, queueUrl), nil

	case "sync":
		return nil, fmt.Errorf("sync driver not yet implemented")

	default:
		return nil, fmt.Errorf("unsupported queue connection: %s", cfg.Queue.Connection)
	}
}

func init() {
	workerCmd.Flags().StringVar(&queueName, "queue", "default", "Name of the queue to process")
	workerCmd.Flags().IntVar(&concurrency, "workers", 5, "Number of concurrent workers")

	root.GetRoot().AddCommand(workerCmd)
}
