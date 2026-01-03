package worker

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/pixelvide/laravel-go/pkg/queue"
)

// Worker manages the processing of jobs
type Worker struct {
	Driver         queue.Driver
	FailedProvider queue.FailedJobProvider
	QueueName      string
	Concurrency    int
	wg             sync.WaitGroup
	quit           chan struct{}
}

// NewWorker creates a new worker instance
func NewWorker(driver queue.Driver, failedProvider queue.FailedJobProvider, queueName string, concurrency int) *Worker {
	return &Worker{
		Driver:         driver,
		FailedProvider: failedProvider,
		QueueName:      queueName,
		Concurrency:    concurrency,
		quit:           make(chan struct{}),
	}
}

// Run starts the worker pool
func (w *Worker) Run(ctx context.Context) {
	for i := 0; i < w.Concurrency; i++ {
		w.wg.Add(1)
		go w.processLoop(ctx, i)
	}
	w.wg.Wait()
}

func (w *Worker) processLoop(ctx context.Context, id int) {
	defer w.wg.Done()
	log.Printf("Worker %d started processing queue: %s", id, w.QueueName)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.quit:
			return
		default:
			// Pop a job
			job, err := w.Driver.Pop(ctx, w.QueueName)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					return
				}
				log.Printf("Worker %d: Error popping job: %v", id, err)
				// Sleep a bit to avoid tight loop on error
				time.Sleep(time.Second)
				continue
			}

			// Process the job
			w.handleJob(ctx, job)
		}
	}
}

func (w *Worker) handleJob(ctx context.Context, job *queue.Job) {
	var payload queue.LaravelJob
	if err := json.Unmarshal(job.Body, &payload); err != nil {
		log.Printf("Error unmarshalling job: %v. Body: %s", err, string(job.Body))
		// If we can't parse it, we probably can't process it.
		// In a real system we might move to failed jobs.
		return
	}

	handler, err := queue.GetHandler(payload.DisplayName)
	if err != nil {
		log.Printf("No handler found for job: %s", payload.DisplayName)
		// TODO: Handle unregistered jobs (maybe fail them?)
		return
	}

	// Attempt to unserialize PHP command if present
	unserialized, _ := queue.UnserializeCommand(payload.Data)

	// Populate job details
	job.Payload = &payload
	job.UnserializedData = unserialized

	// Execute handler
	var jobCtx context.Context
	var cancel context.CancelFunc

	if payload.Timeout != nil && *payload.Timeout > 0 {
		jobCtx, cancel = context.WithTimeout(ctx, time.Duration(*payload.Timeout)*time.Second)
	} else {
		jobCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	err = handler(jobCtx, job)
	if err != nil {
		log.Printf("Job %s failed: %v", payload.DisplayName, err)
		w.handleFailure(ctx, payload, err)
	} else {
		// Job success
		if ackErr := w.Driver.Ack(ctx, job); ackErr != nil {
			log.Printf("Error acknowledging job %s: %v", payload.DisplayName, ackErr)
		}
	}
}

func (w *Worker) handleFailure(ctx context.Context, payload queue.LaravelJob, err error) {
	// Increment attempts
	payload.Attempts++

	maxTries := 1 // default
	if payload.MaxTries != nil {
		maxTries = *payload.MaxTries
	}

	if payload.Attempts < maxTries {
		log.Printf("Retrying job %s (Attempt %d/%d)", payload.DisplayName, payload.Attempts, maxTries)

		// Serialize back to JSON
		body, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			log.Printf("Error marshalling job for retry: %v", marshalErr)
			return
		}

		// Push back to queue
		// Note: In a real Laravel system, we would respect 'backoff' (delayed queue).
		// Here we just push back to the main list (immediate retry) for simplicity,
		// or we could implement a sleep if we are blocking the worker (but that blocks the worker).
		// Ideally the driver supports 'Release(..., delay)'.
		// For MVP, we simply RPUSH (put at end of queue).
		if pushErr := w.Driver.Push(ctx, w.QueueName, body); pushErr != nil {
			log.Printf("Error pushing job back to queue: %v", pushErr)
		}
	} else {
		log.Printf("Job %s failed permanently after %d attempts", payload.DisplayName, payload.Attempts)

		// Serialize payload
		body, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			log.Printf("Error marshalling job for failure: %v", marshalErr)
			return
		}

		if w.FailedProvider != nil {
			// Using "redis" (or driver name) as connection name is a simplification.
			// Ideally we know the connection name from config.
			if failErr := w.FailedProvider.Log(ctx, "redis", w.QueueName, body, err.Error()); failErr != nil {
				log.Printf("Error logging failed job: %v", failErr)
			}
		} else {
			log.Printf("No failed job provider configured. Job lost: %s", payload.DisplayName)
		}
	}
}
