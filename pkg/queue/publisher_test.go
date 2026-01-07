package queue

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDriver is a mock implementation of the Driver interface
type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) Pop(ctx context.Context, queueName string) (*Job, error) {
	args := m.Called(ctx, queueName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Job), args.Error(1)
}

func (m *MockDriver) Push(ctx context.Context, queueName string, body []byte) error {
	args := m.Called(ctx, queueName, body)
	return args.Error(0)
}

func (m *MockDriver) Ack(ctx context.Context, job *Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

func TestPublisher_Dispatch(t *testing.T) {
	mockDriver := new(MockDriver)
	publisher := NewPublisher(mockDriver)

	jobName := "App\\Jobs\\ProcessPodcast"
	args := map[string]interface{}{
		"podcastId": 123,
		"title":     "Go & Laravel",
	}

	mockDriver.On("Push", mock.Anything, "default", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		body := args.Get(2).([]byte)

		var job LaravelJob
		err := json.Unmarshal(body, &job)
		assert.NoError(t, err)

		assert.NotEmpty(t, job.UUID)
		assert.Equal(t, "App\\Jobs\\ProcessPodcast", job.DisplayName)
		assert.Equal(t, "Illuminate\\Queue\\CallQueuedHandler@call", job.Job)

		var dataMap map[string]interface{}
		err = json.Unmarshal(job.Data, &dataMap)
		assert.NoError(t, err)

		assert.Equal(t, "App\\Jobs\\ProcessPodcast", dataMap["commandName"])

		commandStr := dataMap["command"].(string)
		// Check that command string contains the properties
		// Expected: O:23:"App\Jobs\ProcessPodcast":2:{s:9:"podcastId";i:123;s:5:"title";s:12:"Go & Laravel";}
		// Order might vary, so check for containment
		assert.Contains(t, commandStr, "App\\Jobs\\ProcessPodcast")
		assert.Contains(t, commandStr, `"podcastId";i:123;`)
		assert.Contains(t, commandStr, `"title";s:12:"Go & Laravel";`)
	})

	err := publisher.Dispatch(context.Background(), jobName, args)
	assert.NoError(t, err)

	mockDriver.AssertExpectations(t)
}

func TestPublisher_DispatchToQueue(t *testing.T) {
	mockDriver := new(MockDriver)
	publisher := NewPublisher(mockDriver)

	jobName := "App\\Jobs\\SendEmail"
	args := map[string]interface{}{
		"email": "test@example.com",
	}

	queueName := "emails"

	mockDriver.On("Push", mock.Anything, queueName, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		body := args.Get(2).([]byte)

		var job LaravelJob
		err := json.Unmarshal(body, &job)
		assert.NoError(t, err)
		assert.Equal(t, jobName, job.DisplayName)
	})

	err := publisher.DispatchToQueue(context.Background(), queueName, jobName, args)
	assert.NoError(t, err)

	mockDriver.AssertExpectations(t)
}
