package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/pixelvide/laravel-go/pkg/queue"
)

// MockDriver implements queue.Driver for testing
type MockDriver struct {
	Queue    []queue.Job
	Pushed   []queue.Job
	FailCall []queue.Job
}

func (m *MockDriver) Pop(ctx context.Context, queueName string) (*queue.Job, error) {
	if len(m.Queue) > 0 {
		job := m.Queue[0]
		m.Queue = m.Queue[1:]
		return &job, nil
	}
	// Simulate blocking or return error to stop worker loop in test
	return nil, context.DeadlineExceeded
}

func (m *MockDriver) Push(ctx context.Context, queueName string, body []byte) error {
	m.Pushed = append(m.Pushed, queue.Job{Body: body})
	return nil
}

func (m *MockDriver) Fail(ctx context.Context, queueName string, body []byte, err error) error {
	m.FailCall = append(m.FailCall, queue.Job{Body: body})
	return nil
}

func TestWorker_Run_Success(t *testing.T) {
	// Setup Registry
	jobName := "TestJob"
	handled := false
	queue.Register(jobName, func(ctx context.Context, job *queue.Job) error {
		handled = true
		return nil
	})

	// Create Job Payload
	payload := queue.LaravelJob{
		UUID:        "123",
		DisplayName: jobName,
		Attempts:    0,
	}
	body, _ := json.Marshal(payload)

	// Setup Mock Driver
	driver := &MockDriver{
		Queue: []queue.Job{{Body: body}},
	}

	// Create Worker
	w := NewWorker(driver, "default", 1)

	// Run Worker for a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	w.Run(ctx)

	if !handled {
		t.Errorf("Expected handler to be called")
	}
}

func TestWorker_Run_Retry(t *testing.T) {
	// Setup Registry
	jobName := "RetryJob"
	calls := 0
	queue.Register(jobName, func(ctx context.Context, job *queue.Job) error {
		calls++
		return errors.New("failed")
	})

	// Create Job Payload with MaxTries = 2
	maxTries := 2
	payload := queue.LaravelJob{
		UUID:        "456",
		DisplayName: jobName,
		Attempts:    0,
		MaxTries:    &maxTries,
	}
	body, _ := json.Marshal(payload)

	// Setup Mock Driver
	driver := &MockDriver{
		Queue: []queue.Job{{Body: body}},
	}

	// Create Worker
	w := NewWorker(driver, "default", 1)

	// Run Worker
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	w.Run(ctx)

	// Check if pushed back
	if len(driver.Pushed) != 1 {
		t.Errorf("Expected 1 pushed job for retry, got %d", len(driver.Pushed))
	} else {
		// Verify attempts incremented
		var retryPayload queue.LaravelJob
		json.Unmarshal(driver.Pushed[0].Body, &retryPayload)
		if retryPayload.Attempts != 1 {
			t.Errorf("Expected attempts to be 1, got %d", retryPayload.Attempts)
		}
	}
}
