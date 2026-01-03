package config

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSConfig holds configuration for SQS connection
type SQSConfig struct {
	Region    string
	QueueUrl  string
	Profile   string // Optional AWS profile
}

// LoadSQSClient loads an SQS client from config
func LoadSQSClient(ctx context.Context, cfg SQSConfig) (*sqs.Client, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}
	if cfg.Profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(cfg.Profile))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return sqs.NewFromConfig(awsCfg), nil
}
