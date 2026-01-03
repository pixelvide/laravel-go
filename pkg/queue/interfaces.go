package queue

import (
	"context"
)

// Job represents a generic job retrieved from the queue
type Job struct {
	ID   string
	Body []byte
}

// Handler is the function signature for processing a job's payload
type Handler func(ctx context.Context, payload []byte) error

// Driver defines the interface for queue backends
type Driver interface {
	// Pop retrieves a job from the queue. It should block until a job is available.
	Pop(ctx context.Context, queueName string) (*Job, error)
	// Push adds a job payload to the queue
	Push(ctx context.Context, queueName string, body []byte) error
	// Fail moves a job to the failed storage
	Fail(ctx context.Context, queueName string, body []byte, err error) error
}
