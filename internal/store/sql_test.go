package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

func TestDB_Save_Success(t *testing.T) {
	db, mock, ctx := newTestDB(t)
	defer db.Close()

	metric := models.Metrics{
		ID:    "m1",
		MType: models.Counter,
		Delta: ptrInt64(10),
	}

	rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow(metric.ID, metric.MType, *metric.Delta, nil)
	mock.ExpectQuery("INSERT INTO metrics").WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).WillReturnRows(rows)

	saved, err := db.Save(ctx, &metric)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if saved.ID != metric.ID || *saved.Delta != *metric.Delta {
		t.Fatalf("unexpected saved metric: %+v", saved)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestDB_SaveAll_Success(t *testing.T) {
	db, mock, ctx := newTestDB(t)
	defer db.Close()

	metrics := []models.Metrics{
		{ID: "m1", MType: models.Counter, Delta: ptrInt64(10)},
		{ID: "m2", MType: models.Gauge, Value: ptrFloat64(3.14)},
	}

	// Begin transaction
	mock.ExpectBegin()
	stmt := mock.ExpectPrepare("INSERT INTO metrics")
	for _, metric := range metrics {
		stmt.ExpectExec().WithArgs(metric.ID, metric.MType, metric.Delta, metric.Value).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	if err := db.SaveAll(ctx, metrics); err != nil {
		t.Fatalf("SaveAll failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestDB_Get_Success(t *testing.T) {
	db, mock, ctx := newTestDB(t)
	defer db.Close()

	metric := models.Metrics{
		ID:    "m1",
		MType: models.Counter,
		Delta: ptrInt64(10),
	}

	rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
		AddRow(metric.ID, metric.MType, *metric.Delta, nil)
	mock.ExpectQuery("SELECT \\* FROM metrics WHERE id=.*").WithArgs(metric.ID, metric.MType).WillReturnRows(rows)

	got, err := db.Get(ctx, &metric)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != metric.ID || *got.Delta != *metric.Delta {
		t.Fatalf("unexpected metric returned: %+v", got)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestDB_GetAll_Success(t *testing.T) {
	db, mock, ctx := newTestDB(t)
	defer db.Close()

	metrics := []models.Metrics{
		{ID: "m1", MType: models.Counter, Delta: ptrInt64(10)},
		{ID: "m2", MType: models.Gauge, Value: ptrFloat64(3.14)},
	}

	rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"})
	for _, m := range metrics {
		rows.AddRow(m.ID, m.MType, m.Delta, m.Value)
	}
	mock.ExpectQuery("SELECT \\* FROM metrics;").WillReturnRows(rows)

	got, err := db.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(got) != len(metrics) {
		t.Fatalf("expected %d metrics, got %d", len(metrics), len(got))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestDB_Save_UnsupportedMetricType(t *testing.T) {
	db, _, ctx := newTestDB(t)
	defer db.Close()

	metric := models.Metrics{
		ID:    "m1",
		MType: "unsupported",
	}

	_, err := db.Save(ctx, &metric)
	if err == nil {
		t.Fatalf("expected error for unsupported metric type")
	}
}

func TestDB_SaveAll_ErrorOnExec(t *testing.T) {
	db, mock, ctx := newTestDB(t)
	defer db.Close()

	metrics := []models.Metrics{
		{ID: "m1", MType: models.Counter, Delta: ptrInt64(10)},
	}

	mock.ExpectBegin()
	stmt := mock.ExpectPrepare("INSERT INTO metrics")
	stmt.ExpectExec().WillReturnError(errors.New("exec error"))
	mock.ExpectRollback()

	err := db.SaveAll(ctx, metrics)
	if err == nil || err.Error() != "exec error" {
		t.Fatalf("expected exec error, got %v", err)
	}
}

func newTestDB(t *testing.T) (*DB, sqlmock.Sqlmock, context.Context) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	logger := zerolog.Nop()
	testDB := &DB{
		DB:                 db,
		errorClassificator: NewPostgresErrorClassifier(),
		logger:             &logger,
		maxAttempts:        1, // чтобы не тормозить тесты retry
		retryIntervals:     map[int]time.Duration{1: 0},
	}

	return testDB, mock, context.Background()
}
