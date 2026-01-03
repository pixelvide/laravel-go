package queue

import (
	"context"
)

// Job represents a generic job retrieved from the queue
type Job struct {
	ID               string
	Body             []byte
	Payload          *LaravelJob // The parsed JSON envelope
	UnserializedData any         // The unserialized PHP command properties (if applicable)
}

// GetArg retrieves a property from the unserialized PHP command data.
// It is useful for accessing public/protected properties of the Laravel job class.
func (j *Job) GetArg(key string) any {
	if j.UnserializedData == nil {
		return nil
	}
	return GetPHPProperty(j.UnserializedData, key)
}

// Handler is the function signature for processing a job
type Handler func(ctx context.Context, job *Job) error

// Driver defines the interface for queue backends
type Driver interface {
	// Pop retrieves a job from the queue. It should block until a job is available.
	Pop(ctx context.Context, queueName string) (*Job, error)
	// Push adds a job payload to the queue
	Push(ctx context.Context, queueName string, body []byte) error
	// Ack acknowledges that the job has been processed and can be removed
	Ack(ctx context.Context, job *Job) error
}
