package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pixelvide/laravel-go/pkg/queue"
)

type SQSDriver struct {
	client   *sqs.Client
	queueUrl string
}

// NewSQSDriver creates a new SQS driver
func NewSQSDriver(client *sqs.Client, queueUrl string) *SQSDriver {
	return &SQSDriver{
		client:   client,
		queueUrl: queueUrl,
	}
}

// Pop retrieves a job from SQS
func (s *SQSDriver) Pop(ctx context.Context, queueName string) (*queue.Job, error) {
	// Note: queueName argument is ignored if queueUrl is hardcoded in driver,
	// or we could treat queueName as the Queue URL if dynamic.
	// For now, using the configured queueUrl.

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueUrl),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20, // Long polling
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
	}

	resp, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(resp.Messages) == 0 {
		// No messages, return nil (or error depending on expected behavior, usually nil or timeout)
		return nil, context.DeadlineExceeded
	}

	msg := resp.Messages[0]

	// ID is ReceiptHandle (needed for deleting)
	id := ""
	if msg.ReceiptHandle != nil {
		id = *msg.ReceiptHandle
	}

	body := []byte("")
	if msg.Body != nil {
		body = []byte(*msg.Body)
	}

	return &queue.Job{
		ID:   id,
		Body: body,
	}, nil
}

// Push adds a job to SQS
func (s *SQSDriver) Push(ctx context.Context, queueName string, body []byte) error {
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueUrl),
		MessageBody: aws.String(string(body)),
	}

	_, err := s.client.SendMessage(ctx, input)
	return err
}

// Ack deletes the job from SQS
func (s *SQSDriver) Ack(ctx context.Context, job *queue.Job) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueUrl),
		ReceiptHandle: aws.String(job.ID),
	}

	_, err := s.client.DeleteMessage(ctx, input)
	return err
}
