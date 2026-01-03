package queue

import (
	"context"
)

// FailedJobProvider defines the interface for logging failed jobs
type FailedJobProvider interface {
	// Log records a failed job
	Log(ctx context.Context, connection string, queue string, payload []byte, exception string) error
}
