package queue

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/yvasiyarov/php_session_decoder/php_serialize"
)

// Publisher handles dispatching jobs to the queue
type Publisher struct {
	driver Driver
}

// NewPublisher creates a new Publisher instance
func NewPublisher(driver Driver) *Publisher {
	return &Publisher{driver: driver}
}

// Dispatch pushes a new job to the queue
// jobName is the Laravel job class name (e.g., "App\Jobs\ProcessPodcast")
// args is a map of public properties to set on the job object
func (p *Publisher) Dispatch(ctx context.Context, jobName string, args map[string]interface{}) error {
	// 1. Generate UUID
	id := uuid.New().String()

	// 2. Serialize the job object
	// Create a PHP object representing the job class
	phpObj := php_serialize.NewPhpObject(jobName)

	// Set properties
	for key, value := range args {
		// We assume these are public properties
		phpObj.SetPublic(key, value)
	}

	encoder := php_serialize.NewSerializer()
	serializedCommand, err := encoder.Encode(phpObj)
	if err != nil {
		return err
	}

	// 3. Construct the payload data (commandName + command)
	payloadData := map[string]interface{}{
		"commandName": jobName,
		"command":     serializedCommand,
	}

	dataBytes, err := json.Marshal(payloadData)
	if err != nil {
		return err
	}

	// 4. Construct the LaravelJob payload
	laravelJob := LaravelJob{
		UUID:        id,
		DisplayName: jobName,
		Job:         "Illuminate\\Queue\\CallQueuedHandler@call",
		Data:        dataBytes,
	}

	body, err := json.Marshal(laravelJob)
	if err != nil {
		return err
	}

	// 5. Push to the queue
	// Use "default" queue for now, or we could add queue name to Dispatch arguments
	// TODO: Allow specifying queue name
	return p.driver.Push(ctx, "default", body)
}

// DispatchToQueue pushes a new job to a specific queue
func (p *Publisher) DispatchToQueue(ctx context.Context, queueName string, jobName string, args map[string]interface{}) error {
	// 1. Generate UUID
	id := uuid.New().String()

	// 2. Serialize the job object
	phpObj := php_serialize.NewPhpObject(jobName)
	for key, value := range args {
		phpObj.SetPublic(key, value)
	}

	encoder := php_serialize.NewSerializer()
	serializedCommand, err := encoder.Encode(phpObj)
	if err != nil {
		return err
	}

	// 3. Construct the payload data
	payloadData := map[string]interface{}{
		"commandName": jobName,
		"command":     serializedCommand,
	}

	dataBytes, err := json.Marshal(payloadData)
	if err != nil {
		return err
	}

	// 4. Construct the LaravelJob payload
	laravelJob := LaravelJob{
		UUID:        id,
		DisplayName: jobName,
		Job:         "Illuminate\\Queue\\CallQueuedHandler@call",
		Data:        dataBytes,
	}

	body, err := json.Marshal(laravelJob)
	if err != nil {
		return err
	}

	// 5. Push to the specified queue
	return p.driver.Push(ctx, queueName, body)
}
