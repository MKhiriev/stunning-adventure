package observer

import (
	"context"
	"errors"
	"testing"

	"github.com/MKhiriev/stunning-adventure/models"
)

type mockAuditAdapter struct {
	send func(ctx context.Context, event models.AuditEvent) error
}

func (m *mockAuditAdapter) SendEvent(ctx context.Context, event models.AuditEvent) error {
	if m.send != nil {
		return m.send(ctx, event)
	}
	return nil
}

func TestNewRemoteServerObserver_CreatesObserver(t *testing.T) {
	mockAdapter := &mockAuditAdapter{}
	obs := NewRemoteServerObserver(mockAdapter)

	if obs == nil {
		t.Fatal("expected observer, got nil")
	}

	if obs.Name() != "remote server observer" {
		t.Fatalf("unexpected name: %s", obs.Name())
	}
}

func TestRemoteServerObserver_Update_Success(t *testing.T) {
	called := false
	mockAdapter := &mockAuditAdapter{
		send: func(ctx context.Context, event models.AuditEvent) error {
			called = true
			return nil
		},
	}

	obs := NewRemoteServerObserver(mockAdapter)
	event := mustNewAuditEvent(t, "127.0.0.1", "m1")

	err := obs.Update(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !called {
		t.Fatal("expected SendEvent to be called")
	}
}

func TestRemoteServerObserver_Update_Error(t *testing.T) {
	mockAdapter := &mockAuditAdapter{
		send: func(ctx context.Context, event models.AuditEvent) error {
			return errors.New("send failed")
		},
	}

	obs := NewRemoteServerObserver(mockAdapter)
	event := mustNewAuditEvent(t, "127.0.0.1", "m1")

	err := obs.Update(context.Background(), event)
	if err == nil {
		t.Fatal("expected error from Update, got nil")
	}

	if err.Error() != "send failed" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestRemoteServerObserver_Name(t *testing.T) {
	mockAdapter := &mockAuditAdapter{}
	obs := NewRemoteServerObserver(mockAdapter)

	if obs.Name() != "remote server observer" {
		t.Fatalf("unexpected name: %s", obs.Name())
	}
}
