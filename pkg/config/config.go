package config

// Config holds the global application configuration
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Queue    QueueConfig
	Cache    CacheConfig
	Mail     MailConfig
}

// AppConfig maps to APP_* variables
type AppConfig struct {
	Name string `env:"APP_NAME" envDefault:"Laravel"`
	Env  string `env:"APP_ENV" envDefault:"production"`
	Key  string `env:"APP_KEY"`
}

// DatabaseConfig maps to DB_* variables
type DatabaseConfig struct {
	Connection string `env:"DB_CONNECTION" envDefault:"mysql"`
	Host       string `env:"DB_HOST" envDefault:"127.0.0.1"`
	Port       string `env:"DB_PORT" envDefault:"3306"`
	Database   string `env:"DB_DATABASE" envDefault:"forge"`
	Username   string `env:"DB_USERNAME" envDefault:"forge"`
	Password   string `env:"DB_PASSWORD" envDefault:""`
}

// RedisConfig maps to REDIS_* variables
type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
	CacheDB  int    `env:"REDIS_CACHE_DB" envDefault:"1"`
}

// QueueConfig maps to QUEUE_* variables
type QueueConfig struct {
	Connection string `env:"QUEUE_CONNECTION" envDefault:"sync"`
	Queue      string `env:"QUEUE_QUEUE" envDefault:"default"`
}

// CacheConfig maps to CACHE_* variables
type CacheConfig struct {
	Store string `env:"CACHE_STORE" envDefault:"file"` // file, array, database, redis, memcached, dynamo
}

// MailConfig maps to MAIL_* variables
type MailConfig struct {
	Mailer       string `env:"MAIL_MAILER" envDefault:"smtp"`
	Host         string `env:"MAIL_HOST" envDefault:"mailhog"`
	Port         string `env:"MAIL_PORT" envDefault:"1025"`
	Username     string `env:"MAIL_USERNAME"`
	Password     string `env:"MAIL_PASSWORD"`
	Encryption   string `env:"MAIL_ENCRYPTION" envDefault:"tls"`
	FromAddress  string `env:"MAIL_FROM_ADDRESS"`
	FromName     string `env:"MAIL_FROM_NAME"`
}
