package schedule

import (
	"context"
	"time"
)

// LockProvider defines the interface for distributed locks
type LockProvider interface {
	// GetLock attempts to acquire a lock with a given name and duration.
	// Returns true if acquired, false otherwise.
	GetLock(ctx context.Context, name string, duration time.Duration) (bool, error)
	// ReleaseLock releases the lock.
	ReleaseLock(ctx context.Context, name string) error
}
