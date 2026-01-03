package database

import (
	"context"
	"database/sql"
	"time"
)

// DatabaseFailedJobProvider implements queue.FailedJobProvider using a SQL database
type DatabaseFailedJobProvider struct {
	db    *sql.DB
	table string
}

// NewDatabaseFailedJobProvider creates a new provider
func NewDatabaseFailedJobProvider(db *sql.DB, tableName string) *DatabaseFailedJobProvider {
	if tableName == "" {
		tableName = "failed_jobs"
	}
	return &DatabaseFailedJobProvider{
		db:    db,
		table: tableName,
	}
}

// Log records a failed job to the database
func (p *DatabaseFailedJobProvider) Log(ctx context.Context, connection string, queue string, payload []byte, exception string) error {
	query := `
		INSERT INTO ` + p.table + ` (connection, queue, payload, exception, failed_at)
		VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := p.db.ExecContext(ctx, query, connection, queue, payload, exception, now)
	return err
}
