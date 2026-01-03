package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/pixelvide/laravel-go/pkg/queue"
)

// DatabaseDriver implements queue.Driver for SQL databases
type DatabaseDriver struct {
	db     *sql.DB
	table  string
	driver string
	mu     sync.RWMutex
}

// NewDatabaseDriver creates a new database driver
func NewDatabaseDriver(cfg config.DatabaseConfig, db *sql.DB) *DatabaseDriver {
	tableName := cfg.Table
	if tableName == "" {
		tableName = "jobs"
	}
	return &DatabaseDriver{
		db:     db,
		table:  tableName,
		driver: cfg.Driver,
	}
}

func (d *DatabaseDriver) rebind(query string) string {
	d.mu.RLock()
	driver := d.driver
	d.mu.RUnlock()

	if driver != "postgres" && driver != "pq" && driver != "pgsql" {
		return query
	}

	// Replace ? with $1, $2, etc.
	parts := strings.Split(query, "?")
	if len(parts) == 1 {
		return query
	}

	var builder strings.Builder
	for i, part := range parts {
		builder.WriteString(part)
		if i < len(parts)-1 {
			builder.WriteString("$" + strconv.Itoa(i+1))
		}
	}
	return builder.String()
}

// Pop retrieves a job from the database
// Note: This is a simplified implementation. A production-ready version
// should handle "reserved" state, row locking (FOR UPDATE SKIP LOCKED),
// and specific SQL dialects (Postgres vs MySQL).
func (d *DatabaseDriver) Pop(ctx context.Context, queueName string) (*queue.Job, error) {
	// Simple polling loop since SQL doesn't block like Redis
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Attempt to pop a job
			job, err := d.popJob(ctx, queueName)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				return nil, err
			}
			return job, nil
		}
	}
}

func (d *DatabaseDriver) popJob(ctx context.Context, queueName string) (*queue.Job, error) {
	// Start transaction
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Find available job
	// Laravel jobs table usually has: id, queue, payload, attempts, reserved_at, available_at
	query := fmt.Sprintf(`
		SELECT id, payload
		FROM %s
		WHERE queue = ?
		AND (reserved_at IS NULL OR reserved_at <= ?)
		AND available_at <= ?
		ORDER BY id ASC
		LIMIT 1 FOR UPDATE`, d.table) // Note: FOR UPDATE blocks if not using SKIP LOCKED (Postgres feature)

	query = d.rebind(query)

	// For compatibility/simplicity, we assume standard SQL.
	// In high concurrency, this might lock.
	// We use current timestamp for checks
	now := time.Now().Unix()

	var id int64
	var payload []byte

	err = tx.QueryRowContext(ctx, query, queueName, now, now).Scan(&id, &payload)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			// Auto-detect PostgreSQL driver if not configured
			if strings.Contains(err.Error(), "pq:") || strings.Contains(err.Error(), "syntax error at or near \"AND\"") {
				d.mu.RLock()
				currentDriver := d.driver
				d.mu.RUnlock()

				if currentDriver != "postgres" && currentDriver != "pq" && currentDriver != "pgsql" {
					log.Printf("[DatabaseDriver] Detected PostgreSQL error: %v. Switching driver mode to 'postgres'.", err)
					d.mu.Lock()
					d.driver = "postgres"
					d.mu.Unlock()

					// Rollback current transaction and retry
					_ = tx.Rollback()
					return d.popJob(ctx, queueName)
				}
			}
			log.Printf("[DatabaseDriver] Error popping job (Driver=%s): %v. Query: %s", d.driver, err, query)
		}
		return nil, err
	}

	// Delete job (Queue worker style: pop means remove)
	// Laravel usually marks as reserved, then deletes on Ack.
	// Since our interface is Pop-Consume-Done, we can delete here OR mark reserved.
	// To match the Redis implementation (which pops/removes), we delete it.
	// NOTE: If worker crashes, job is lost.
	// Improvement: Mark reserved, delete on completion. But interface is simple Pop.
	deleteQuery := d.rebind(fmt.Sprintf("DELETE FROM %s WHERE id = ?", d.table))
	_, err = tx.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &queue.Job{
		ID:   fmt.Sprintf("%d", id),
		Body: payload,
	}, nil
}

// Push adds a job to the database
func (d *DatabaseDriver) Push(ctx context.Context, queueName string, body []byte) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (queue, payload, attempts, available_at, created_at)
		VALUES (?, ?, 0, ?, ?)`, d.table)

	query = d.rebind(query)

	now := time.Now().Unix()
	_, err := d.db.ExecContext(ctx, query, queueName, body, now, now)
	return err
}

// Ack deletes the job from the database
func (d *DatabaseDriver) Ack(ctx context.Context, job *queue.Job) error {
	// Job was already deleted in Pop, so this is a no-op in the current implementation.
	// If Pop used "mark reserved" strategy, this would delete it.
	return nil
}

// Fail moves a job to the failed_jobs table
func (d *DatabaseDriver) Fail(ctx context.Context, queueName string, body []byte, err error) error {
	// Laravel failed_jobs: uuid, connection, queue, payload, exception, failed_at
	query := `
		INSERT INTO failed_jobs (connection, queue, payload, exception, failed_at)
		VALUES (?, ?, ?, ?, ?)`

	query = d.rebind(query)

	now := time.Now() // failed_at is usually timestamp
	// We use "database" as connection name
	_, dbErr := d.db.ExecContext(ctx, query, "database", queueName, body, err.Error(), now)
	return dbErr
}
