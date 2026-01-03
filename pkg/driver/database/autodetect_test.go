package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pixelvide/laravel-go/pkg/config"
)

func TestPop_PostgresAutoDetection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Initial config is NOT postgres (default/empty)
	cfg := config.DatabaseConfig{Table: "jobs"}
	driver := NewDatabaseDriver(cfg, db)

	// FIRST CALL: Fails with pq: syntax error
	mock.ExpectBegin()
	// Expect query with ? (mysql style)
	query1 := `SELECT id, payload FROM jobs WHERE queue = \? AND \(reserved_at IS NULL OR reserved_at <= \?\) AND available_at <= \? ORDER BY id ASC LIMIT 1 FOR UPDATE`

	// We simulate an error from Postgres
	mock.ExpectQuery(query1).
		WithArgs("default", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("pq: syntax error at or near \"AND\""))

	mock.ExpectRollback()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Run Pop -> Should fail, but trigger auto-detection
	_, err = driver.popJob(ctx, "default")
	if err == nil {
		t.Fatal("Expected error on first pop")
	}

	// Verify driver mode switched
	driver.mu.RLock()
	currentDriver := driver.driver
	driver.mu.RUnlock()
	if currentDriver != "postgres" {
		t.Errorf("Expected driver to switch to postgres, got %s", currentDriver)
	}

	// SECOND CALL: Should use $1 syntax
	mock.ExpectBegin()
	// Expect query with $1 (postgres style)
	query2 := `SELECT id, payload FROM jobs WHERE queue = \$1 AND \(reserved_at IS NULL OR reserved_at <= \$2\) AND available_at <= \$3 ORDER BY id ASC LIMIT 1 FOR UPDATE`

	mock.ExpectQuery(query2).
		WithArgs("default", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payload"}).AddRow(1, []byte("{}")))

	mock.ExpectExec("DELETE FROM jobs WHERE id = \\$1").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// 2. Run Pop -> Should succeed with new syntax
	job, err := driver.popJob(ctx, "default")
	if err != nil {
		t.Errorf("Second pop failed: %v", err)
	}
	if job == nil {
		t.Error("Expected job, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
