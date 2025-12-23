package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MKhiriev/stunning-adventure/internal/service/observer"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type mockAuditAdapter struct {
	called bool
	fail   bool
}

type mockFileObserver struct {
	observer.AuditObserver
	called bool
	fail   bool
}

func TestNewObservers_CreatesObservers(t *testing.T) {
	mockAdapter := &mockAuditAdapter{}
	logger := zerolog.Nop()
	obs := observer.NewObservers("/tmp/audit.log", mockAdapter, &logger)

	if obs.FileObserver == nil {
		t.Fatal("expected FileObserver, got nil")
	}
	if obs.RemoteServerObserver == nil {
		t.Fatal("expected RemoteServerObserver, got nil")
	}
}

func TestNewAuditService_RegisterAndNotifyAll_Success(t *testing.T) {
	mockAdapter := &mockAuditAdapter{}
	fileObs := &mockFileObserver{}
	logger := zerolog.Nop()

	// создаем сервис
	svc := NewAuditService("filePath", mockAdapter, &logger)
	if svc == nil {
		t.Fatal("expected auditService, got nil")
	}

	auditSvc, ok := svc.(*auditService)
	if !ok {
		t.Fatal("expected *auditService type")
	}

	// перезаписываем наблюдателей на мок-файлового
	auditSvc.observers = map[string]observer.AuditObserver{
		"file":   fileObs,
		"remote": observer.NewRemoteServerObserver(mockAdapter),
	}

	event := mustEvent(t)
	err := auditSvc.NotifyAll(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем, что оба observer вызвались
	if !fileObs.called {
		t.Fatal("file observer was not called")
	}
	if !mockAdapter.called {
		t.Fatal("remote observer was not called")
	}
}

func TestNewAuditService_NotifyAll_ErrorsAggregated(t *testing.T) {
	mockAdapter := &mockAuditAdapter{fail: true}
	fileObs := &mockFileObserver{fail: true}
	logger := zerolog.Nop()

	svc := &auditService{
		observers: map[string]observer.AuditObserver{
			"file":   fileObs,
			"remote": observer.NewRemoteServerObserver(mockAdapter),
		},
		logger: &logger,
	}

	event := mustEvent(t)
	err := svc.NotifyAll(context.Background(), event)
	if err == nil {
		t.Fatal("expected aggregated error, got nil")
	}

	multiErr, ok := err.(observer.MultiObserverError)
	if !ok {
		t.Fatal("expected MultiObserverError type")
	}

	if len(multiErr.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(multiErr.Errors))
	}
}

func TestAuditService_RegisterAndDeregister(t *testing.T) {
	logger := zerolog.Nop()
	svc := &auditService{
		observers: make(map[string]observer.AuditObserver),
		logger:    &logger,
	}

	mockObs := &mockFileObserver{}

	// Регистрация
	err := svc.Register(mockObs)
	if err != nil {
		t.Fatalf("unexpected error registering observer: %v", err)
	}

	if len(svc.observers) != 1 {
		t.Fatal("observer not registered")
	}

	// Повторная регистрация того же observer
	err = svc.Register(mockObs)
	if err != nil {
		t.Fatalf("unexpected error re-registering observer: %v", err)
	}

	// Удаление
	err = svc.Deregister(mockObs)
	if err != nil {
		t.Fatalf("unexpected error deregistering observer: %v", err)
	}

	if len(svc.observers) != 0 {
		t.Fatal("observer not deregistered")
	}

	// Удаление несуществующего observer
	err = svc.Deregister(mockObs)
	if err == nil {
		t.Fatal("expected error deregistering non-registered observer")
	}
}

func (m *mockAuditAdapter) SendEvent(ctx context.Context, event models.AuditEvent) error {
	m.called = true
	if m.fail {
		return errors.New("adapter failed")
	}
	return nil
}

func (f *mockFileObserver) Update(ctx context.Context, event models.AuditEvent) error {
	f.called = true
	if f.fail {
		return errors.New("file write failed")
	}
	return nil
}

func (f *mockFileObserver) Name() string {
	return "mock file observer"
}

// helper to create a simple audit event
func mustEvent(t *testing.T) models.AuditEvent {
	t.Helper()
	ev, err := models.NewAuditEvent("127.0.0.1", time.Now(), "m1", "m2")
	if err != nil {
		t.Fatal(err)
	}
	return ev
}
