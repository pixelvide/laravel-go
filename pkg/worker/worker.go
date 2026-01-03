package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/pixelvide/laravel-go/pkg/queue"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Worker manages the processing of jobs
type Worker struct {
	Driver         queue.Driver
	FailedProvider queue.FailedJobProvider
	QueueName      string
	Concurrency    int
	Tracer         trace.Tracer
	wg             sync.WaitGroup
	quit           chan struct{}
}

// NewWorker creates a new worker instance
func NewWorker(driver queue.Driver, failedProvider queue.FailedJobProvider, queueName string, concurrency int, tracer trace.Tracer) *Worker {
	if tracer == nil {
		tracer = otel.Tracer("worker")
	}
	return &Worker{
		Driver:         driver,
		FailedProvider: failedProvider,
		QueueName:      queueName,
		Concurrency:    concurrency,
		Tracer:         tracer,
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
	log.Info().Int("worker_id", id).Str("queue", w.QueueName).Msg("Worker started processing queue")

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
				log.Error().Err(err).Int("worker_id", id).Msg("Error popping job")
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
		log.Error().Err(err).Str("body", string(job.Body)).Msg("Error unmarshalling job")
		// If we can't parse it, we probably can't process it.
		// In a real system we might move to failed jobs.
		return
	}

	// Start Trace
	ctx, span := w.Tracer.Start(ctx, "process_job")
	defer span.End()

	// Extract TraceID and Setup Logger
	traceID := span.SpanContext().TraceID().String()
	logger := log.With().
		Str("trace_id", traceID).
		Str("job_uuid", payload.UUID).
		Str("job_name", payload.DisplayName).
		Logger()

	// Inject logger into context
	ctx = logger.WithContext(ctx)

	handler, err := queue.GetHandler(payload.DisplayName)
	if err != nil {
		logger.Error().Str("job_name", payload.DisplayName).Msg("No handler found for job")
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
		logger.Error().Err(err).Msg("Job failed")
		w.handleFailure(ctx, payload, err)
	} else {
		// Job success
		if ackErr := w.Driver.Ack(ctx, job); ackErr != nil {
			logger.Error().Err(ackErr).Msg("Error acknowledging job")
		} else {
			logger.Info().Msg("Job processed successfully")
		}
	}
}

func (w *Worker) handleFailure(ctx context.Context, payload queue.LaravelJob, err error) {
	logger := zerolog.Ctx(ctx)

	// Increment attempts
	payload.Attempts++

	maxTries := 1 // default
	if payload.MaxTries != nil {
		maxTries = *payload.MaxTries
	}

	if payload.Attempts < maxTries {
		logger.Info().Int("attempt", payload.Attempts).Int("max_tries", maxTries).Msg("Retrying job")

		// Serialize back to JSON
		body, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			logger.Error().Err(marshalErr).Msg("Error marshalling job for retry")
			return
		}

		// Push back to queue
		// Note: In a real Laravel system, we would respect 'backoff' (delayed queue).
		// Here we just push back to the main list (immediate retry) for simplicity,
		// or we could implement a sleep if we are blocking the worker (but that blocks the worker).
		// Ideally the driver supports 'Release(..., delay)'.
		// For MVP, we simply RPUSH (put at end of queue).
		if pushErr := w.Driver.Push(ctx, w.QueueName, body); pushErr != nil {
			logger.Error().Err(pushErr).Msg("Error pushing job back to queue")
		}
	} else {
		logger.Error().Int("attempts", payload.Attempts).Msg("Job failed permanently")

		// Serialize payload
		body, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			logger.Error().Err(marshalErr).Msg("Error marshalling job for failure")
			return
		}

		if w.FailedProvider != nil {
			// Using "redis" (or driver name) as connection name is a simplification.
			// Ideally we know the connection name from config.
			if failErr := w.FailedProvider.Log(ctx, "redis", w.QueueName, body, err.Error()); failErr != nil {
				logger.Error().Err(failErr).Msg("Error logging failed job")
			}
		} else {
			logger.Error().Msg("No failed job provider configured. Job lost")
		}
	}
}
