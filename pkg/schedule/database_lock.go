package schedule

import (
	"context"
	"database/sql"
	"fmt"
	"hash/crc32"
	"time"
)

// DatabaseLockProvider implements LockProvider using SQL database locks
type DatabaseLockProvider struct {
	db     *sql.DB
	driver string // "mysql" or "postgres"
}

// NewDatabaseLockProvider creates a new database lock provider
func NewDatabaseLockProvider(db *sql.DB, driver string) *DatabaseLockProvider {
	return &DatabaseLockProvider{
		db:     db,
		driver: driver,
	}
}

// GetLock attempts to acquire a lock
func (d *DatabaseLockProvider) GetLock(ctx context.Context, name string, duration time.Duration) (bool, error) {
	if d.driver == "postgres" || d.driver == "pgsql" || d.driver == "pq" {
		return d.getPostgresLock(ctx, name)
	}
	return d.getMySQLLock(ctx, name, duration)
}

// ReleaseLock releases the lock
func (d *DatabaseLockProvider) ReleaseLock(ctx context.Context, name string) error {
	if d.driver == "postgres" || d.driver == "pgsql" || d.driver == "pq" {
		return d.releasePostgresLock(ctx, name)
	}
	return d.releaseMySQLLock(ctx, name)
}

// MySQL Implementation using GET_LOCK
func (d *DatabaseLockProvider) getMySQLLock(ctx context.Context, name string, duration time.Duration) (bool, error) {
	// GET_LOCK(str, timeout) returns 1 if success, 0 if timeout, NULL if error
	// We use timeout 0 to return immediately
	query := "SELECT GET_LOCK(?, 0)"
	var result sql.NullInt64
	err := d.db.QueryRowContext(ctx, query, name).Scan(&result)
	if err != nil {
		return false, err
	}
	if !result.Valid {
		return false, fmt.Errorf("GET_LOCK returned NULL")
	}
	return result.Int64 == 1, nil
}

func (d *DatabaseLockProvider) releaseMySQLLock(ctx context.Context, name string) error {
	// RELEASE_LOCK(str)
	query := "SELECT RELEASE_LOCK(?)"
	var result sql.NullInt64
	err := d.db.QueryRowContext(ctx, query, name).Scan(&result)
	return err
}

// Postgres Implementation using Advisory Locks
func (d *DatabaseLockProvider) getPostgresLock(ctx context.Context, name string) (bool, error) {
	// pg_try_advisory_lock(key)
	// Key must be int64. We hash the string name to int64.
	key := d.hashName(name)
	query := "SELECT pg_try_advisory_lock($1)"
	var success bool
	err := d.db.QueryRowContext(ctx, query, key).Scan(&success)
	if err != nil {
		return false, err
	}
	return success, nil
}

func (d *DatabaseLockProvider) releasePostgresLock(ctx context.Context, name string) error {
	// pg_advisory_unlock(key)
	key := d.hashName(name)
	query := "SELECT pg_advisory_unlock($1)"
	var success bool
	err := d.db.QueryRowContext(ctx, query, key).Scan(&success)
	return err
}

func (d *DatabaseLockProvider) hashName(name string) int64 {
	// Use CRC32 to generate a deterministic integer from string
	// pg_advisory_lock takes bigint (int64)
	return int64(crc32.ChecksumIEEE([]byte(name)))
}
