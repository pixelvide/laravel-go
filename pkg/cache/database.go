package cache

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type DatabaseStore struct {
	db     *sql.DB
	table  string
	driver string // "mysql", "postgres"
}

// NewDatabaseStore creates a new database cache store
// driverName should be "mysql" or "postgres" (or "pgsql")
func NewDatabaseStore(db *sql.DB, table string, driverName string) *DatabaseStore {
	if table == "" {
		table = "cache"
	}
	return &DatabaseStore{db: db, table: table, driver: driverName}
}

func (s *DatabaseStore) quote(identifier string) string {
	if s.driver == "postgres" || s.driver == "pgsql" || s.driver == "pq" {
		return fmt.Sprintf(`"%s"`, identifier)
	}
	// Default to mysql backticks
	return fmt.Sprintf("`%s`", identifier)
}

func (s *DatabaseStore) Get(ctx context.Context, key string) (string, error) {
	// Check expiration
	query := fmt.Sprintf("SELECT value FROM %s WHERE %s = ? AND expiration >= ?", s.table, s.quote("key"))

	now := time.Now().Unix()
	var value string
	err := s.db.QueryRowContext(ctx, query, key, now).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (s *DatabaseStore) Put(ctx context.Context, key string, value string, ttl time.Duration) error {
	expiration := time.Now().Add(ttl).Unix()

	// Simplify: Delete then Insert (Atomic issues but simpler for multi-driver)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete old
	delQuery := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", s.table, s.quote("key"))
	if _, err := tx.ExecContext(ctx, delQuery, key); err != nil {
		return err
	}

	// Insert new
	insQuery := fmt.Sprintf("INSERT INTO %s (%s, value, expiration) VALUES (?, ?, ?)", s.table, s.quote("key"))
	if _, err := tx.ExecContext(ctx, insQuery, key, value, expiration); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *DatabaseStore) Forget(ctx context.Context, key string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", s.table, s.quote("key"))
	_, err := s.db.ExecContext(ctx, query, key)
	return err
}

func (s *DatabaseStore) Flush(ctx context.Context) error {
	query := fmt.Sprintf("DELETE FROM %s", s.table)
	_, err := s.db.ExecContext(ctx, query)
	return err
}
