package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pixelvide/laravel-go/pkg/config"
)

func TestDatabaseDriver_Push(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	driver := NewDatabaseDriver(config.DatabaseConfig{Table: "jobs"}, db)

	queueName := "default"
	body := []byte(`{"job":"Test"}`)

	mock.ExpectExec("INSERT INTO jobs").
		WithArgs(queueName, body, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = driver.Push(context.Background(), queueName, body)
	if err != nil {
		t.Errorf("error was not expected while pushing job: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDatabaseDriver_Pop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	driver := NewDatabaseDriver(config.DatabaseConfig{Table: "jobs"}, db)

	queueName := "default"
	body := []byte(`{"job":"Test"}`)
	// now := time.Now().Unix()

	// Mock transaction begin
	mock.ExpectBegin()

	// Mock Select
	rows := sqlmock.NewRows([]string{"id", "payload"}).AddRow(1, body)
	mock.ExpectQuery("SELECT id, payload FROM jobs").
		WithArgs(queueName, sqlmock.AnyArg(), sqlmock.AnyArg()). // checking for roughly 'now' is hard in mock
		WillReturnRows(rows)

	// Mock Delete
	mock.ExpectExec("DELETE FROM jobs").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock Commit
	mock.ExpectCommit()

	// Use a context with timeout to ensure loop breaks if logic is wrong (though Pop loop checks for success)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// We need to inject the mock logic for Pop.
	// The Pop method has a ticker loop.
	// To test it without waiting too long, we can just run it.
	// Since the select returns immediately, the loop in Pop should catch it on first tick or immediately?
	// The implementation uses `ticker := time.NewTicker(1 * time.Second)`. This will delay test by 1s.
	// We can't easily mock time in standard lib.
	// But it's fine for a test to take 1s.

	job, err := driver.Pop(ctx, queueName)
	if err != nil {
		t.Errorf("error was not expected while popping job: %s", err)
	}

	if job == nil {
		t.Errorf("expected job, got nil")
	}

	if string(job.Body) != string(body) {
		t.Errorf("expected body %s, got %s", body, job.Body)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDatabaseDriver_Fail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	driver := NewDatabaseDriver(config.DatabaseConfig{Table: "jobs"}, db)

	queueName := "default"
	body := []byte(`{"job":"Test"}`)
	failErr := errors.New("something went wrong")

	mock.ExpectExec("INSERT INTO failed_jobs").
		WithArgs("database", queueName, body, failErr.Error(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = driver.Fail(context.Background(), queueName, body, failErr)
	if err != nil {
		t.Errorf("error was not expected while failing job: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
