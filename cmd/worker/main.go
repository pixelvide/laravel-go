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
	w := worker.NewWorker(driver, queueName, concurrency)

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
	w.Run(ctx)
	log.Println("Worker pool stopped.")
}
