package service

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MKhiriev/stunning-adventure/internal/store"
	"github.com/rs/zerolog"
)

func TestPingDBService_Ping_OK(t *testing.T) {
	db, mock, ctx := newTestPingDB(t)

	mock.ExpectPing().WillReturnError(nil)

	logger := zerolog.Nop()
	svc := &PingDBService{
		DB:  db,
		log: &logger,
	}

	err := svc.Ping(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPingDBService_Ping_Error(t *testing.T) {
	db, mock, ctx := newTestPingDB(t)

	pingErr := errors.New("db is down")
	mock.ExpectPing().WillReturnError(pingErr)

	logger := zerolog.Nop()
	svc := &PingDBService{
		DB:  db,
		log: &logger,
	}

	err := svc.Ping(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, pingErr) {
		t.Fatalf("expected error %v, got %v", pingErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestPingDBService_Ping_DBIsNil(t *testing.T) {
	logger := zerolog.Nop()

	svc := &PingDBService{
		DB:  nil,
		log: &logger,
	}

	err := svc.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "DB connection is nil" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewPingDBService(t *testing.T) {
	logger := zerolog.Nop()

	db := &store.DB{}

	svc, err := NewPingDBService(db, &logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func newTestPingDB(t *testing.T) (*store.DB, sqlmock.Sqlmock, context.Context) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	testDB := &store.DB{
		DB: db,
	}

	return testDB, mock, context.Background()
}
