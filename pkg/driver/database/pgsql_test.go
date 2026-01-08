package database

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pixelvide/laravel-go/pkg/config"
)

func TestPop_PgSQLConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Configure driver as pgsql (alias for postgres)
	cfg := config.DatabaseConfig{Connection: "pgsql"}
	driver := NewDatabaseDriver(cfg, db)

	// Expectation
	mock.ExpectBegin()

	// We expect the query with $1, $2, $3 because rebind should have replaced ?
	query := `SELECT id, payload FROM jobs WHERE queue = \$1 AND \(reserved_at IS NULL OR reserved_at <= \$2\) AND available_at <= \$3 ORDER BY id ASC LIMIT 1 FOR UPDATE`

	mock.ExpectQuery(query).
		WithArgs("default", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payload"}).AddRow(1, []byte("{}")))

	// Delete query: DELETE FROM jobs WHERE id = $1
	mock.ExpectExec("DELETE FROM jobs WHERE id = \\$1").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = driver.Pop(ctx, "default")
	if err != nil {
		t.Errorf("Pop failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
