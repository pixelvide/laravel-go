package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/driver/redis"
	"github.com/pixelvide/laravel-go/pkg/queue"
	"github.com/pixelvide/laravel-go/pkg/schedule"
	"github.com/pixelvide/laravel-go/pkg/worker"
)

// ExampleHandler is a sample job handler
func ExampleHandler(ctx context.Context, job *queue.Job) error {
	log.Printf("Processing job: %s, ID: %s", job.Payload.DisplayName, job.Payload.UUID)

	// Example of accessing unserialized PHP data
	if job.UnserializedData != nil {
		// Map properties if needed
		// props := queue.GetPHPProperty(job.UnserializedData, "podcastId")
		log.Printf("Unserialized data present")
	}
	return nil
}

func main() {
	// 1. Configure
	redisConfig := config.RedisConfig{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	queueName := "default"
	concurrency := 5

	// 2. Register Handlers
	// Register a handler for a hypothetical Laravel job "App\Jobs\ProcessPodcast"
	queue.Register("App\\Jobs\\ProcessPodcast", ExampleHandler)

	// 3. Initialize Driver
	driver := redis.NewRedisDriver(redisConfig)

	// 4. Initialize Worker
	// For this example, we aren't setting up a database failed job provider, so we pass nil
	w := worker.NewWorker(driver, nil, queueName, concurrency)

	// 5. Run Worker with Graceful Shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Shutting down worker...")
		cancel()
	}()

	log.Println("Starting worker pool...")

	// Example: Run Scheduler (Optional)
	// go runScheduler(redisConfig)

	w.Run(ctx)
	log.Println("Worker pool stopped.")
}

func runScheduler(redisCfg config.RedisConfig) {
	// Example of starting the scheduler
	redisClient := redis.NewRedisDriver(redisCfg).Client

	lockProvider := schedule.NewRedisLockProvider(redisClient)
	kernel := schedule.NewKernel(lockProvider)

	kernel.Register("* * * * *", func() {
		log.Println("Running scheduled task...")
	}, schedule.OnOneServer("my-scheduled-task"))

	kernel.Run()
}
