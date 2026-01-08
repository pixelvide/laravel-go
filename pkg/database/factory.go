package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/pixelvide/laravel-go/pkg/config"
)

// Factory creates database connections
type Factory struct{}

// NewFactory creates a new Factory
func NewFactory() *Factory {
	return &Factory{}
}

// Connect creates a new database connection based on configuration
func (f *Factory) Connect(cfg config.DatabaseConfig) (*sql.DB, error) {
	var dsn string
	var driverName string

	switch cfg.Connection {
	case "mysql":
		driverName = "mysql"
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	case "pgsql", "postgres":
		driverName = "postgres"
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database)
	default:
		return nil, fmt.Errorf("unsupported database connection: %s", cfg.Connection)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	// Basic configuration
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
