package config

// RedisConfig holds configuration for Redis connection
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// WorkerConfig holds configuration for the worker pool
type WorkerConfig struct {
	QueueName   string
	Concurrency int
	Redis       RedisConfig
}
