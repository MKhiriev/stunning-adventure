package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/stunning-adventure/internal/service/observer"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type mockAuditObserver struct {
	name         string
	updateFn     func(ctx context.Context, event models.AuditEvent) error
	updateCalled bool
}

func TestAuditService_Register_Deregister(t *testing.T) {
	logger := zerolog.Nop()
	svc := &auditService{
		observers: make(map[string]observer.AuditObserver),
		logger:    &logger,
	}

	m := &mockAuditObserver{name: "test"}

	// Register
	if err := svc.Register(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := svc.observers["test"]; !ok {
		t.Fatal("observer was not registered")
	}

	// Deregister
	if err := svc.Deregister(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := svc.observers["test"]; ok {
		t.Fatal("observer was not deregistered")
	}

	// Deregister unregistered
	err := svc.Deregister(m)
	if err == nil {
		t.Fatal("expected error for deregistering unregistered observer")
	}

	// Register nil
	if err := svc.Register(nil); err == nil {
		t.Fatal("expected error for registering nil observer")
	}

	// Deregister nil
	if err := svc.Deregister(nil); err == nil {
		t.Fatal("expected error for deregistering nil observer")
	}
}

func TestAuditService_NotifyAll(t *testing.T) {
	logger := zerolog.Nop()
	m1 := &mockAuditObserver{name: "obs1"}
	m2 := &mockAuditObserver{name: "obs2", updateFn: func(_ context.Context, _ models.AuditEvent) error {
		return errors.New("fail")
	}}

	svc := &auditService{
		observers: map[string]observer.AuditObserver{
			m1.Name(): m1,
			m2.Name(): m2,
		},
		logger: &logger,
	}

	event := models.AuditEvent{IPAddress: "127.0.0.1"}

	// Notify with one failing observer
	err := svc.NotifyAll(context.Background(), event)
	if err == nil {
		t.Fatal("expected MultiObserverError")
	}

	var multiErr observer.MultiObserverError
	ok := errors.As(err, &multiErr)
	if !ok {
		t.Fatal("expected MultiObserverError type")
	}

	if len(multiErr.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(multiErr.Errors))
	}

	// Notify with all OK
	m2.updateFn = func(_ context.Context, _ models.AuditEvent) error { return nil }
	err = svc.NotifyAll(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m1.updateCalled || !m2.updateCalled {
		t.Fatal("Update was not called on all observers")
	}
}

func TestNewAuditService_ReturnsNilWhenNoObservers(t *testing.T) {
	logger := zerolog.Nop()
	svc := NewAuditService("", nil, &logger)
	if svc != nil {
		t.Fatal("expected nil when no observers are provided")
	}
}

func (m *mockAuditObserver) Update(ctx context.Context, event models.AuditEvent) error {
	m.updateCalled = true
	if m.updateFn != nil {
		return m.updateFn(ctx, event)
	}
	return nil
}

func (m *mockAuditObserver) Name() string {
	return m.name
}
