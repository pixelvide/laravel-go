package config

// RedisConfig holds configuration for Redis connection
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// DatabaseConfig holds configuration for SQL database connection
type DatabaseConfig struct {
	Driver string // mysql, postgres
	DSN    string // Data Source Name
	Table  string // jobs table name, default "jobs"
}

// WorkerConfig holds configuration for the worker pool
type WorkerConfig struct {
	QueueName   string
	Concurrency int
	Redis       *RedisConfig
	Database    *DatabaseConfig
}
