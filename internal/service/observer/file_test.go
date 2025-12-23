package observer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

// helper to create a valid AuditEvent
func mustNewAuditEvent(t *testing.T, ip string, names ...string) models.AuditEvent {
	t.Helper()
	ev, err := models.NewAuditEvent(ip, time.Now(), names...)
	if err != nil {
		t.Fatalf("failed to create audit event: %v", err)
	}
	return ev
}

func TestNewFileObserver_SuccessAndUpdateWritesFile(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "audit.log")

	logger := zerolog.Nop()
	observer := NewFileObserver(filePath, &logger)
	if observer == nil {
		t.Fatalf("expected FileObserver, got nil")
	}

	// cast to concrete type to call Name (not necessary but fine)
	fileObserver, ok := observer.(*FileObserver)
	if !ok {
		t.Fatalf("expected *FileObserver, got %T", observer)
	}
	if fileObserver.Name() == "" {
		t.Fatalf("expected non-empty name")
	}

	// Create an event and update (append to file)
	event := mustNewAuditEvent(t, "127.0.0.1", "m1", "m2")
	if err := fileObserver.Update(context.Background(), event); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// read file and ensure JSON of event is present (single line)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// file should contain one JSON line (plus newline)
	line := string(data)
	if len(line) == 0 {
		t.Fatal("file is empty after Update")
	}

	// trim newline and unmarshal to verify correct content
	trimmed := line
	if trimmed[len(trimmed)-1] == '\n' {
		trimmed = trimmed[:len(trimmed)-1]
	}

	var got models.AuditEvent
	if err := json.Unmarshal([]byte(trimmed), &got); err != nil {
		t.Fatalf("invalid JSON in file: %v", err)
	}

	if got.IPAddress != event.IPAddress {
		t.Fatalf("unexpected IPAddress: want %q got %q", event.IPAddress, got.IPAddress)
	}
	if len(got.Metrics) != len(event.Metrics) {
		t.Fatalf("unexpected metrics length: want %d got %d", len(event.Metrics), len(got.Metrics))
	}
}

func TestNewFileObserver_CreateFileError_ReturnsNil(t *testing.T) {
	tmp := t.TempDir()

	// Create a regular file where directory is expected, to make MkdirAll(path.Dir(filePath)) fail.
	blocking := filepath.Join(tmp, "notadir")
	if err := os.WriteFile(blocking, []byte("x"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Now filePath's parent is a regular file -> MkdirAll should fail
	filePath := filepath.Join(blocking, "audit.log")

	logger := zerolog.Nop()
	observer := NewFileObserver(filePath, &logger)
	if observer != nil {
		t.Fatalf("expected nil FileObserver when CreateFile fails, got %T", observer)
	}
}

func TestFileObserver_Update_AppendError(t *testing.T) {
	tmp := t.TempDir()

	// Create a concrete FileObserver but point filePath to a directory.
	// AppendToFile should fail when trying to open a directory for writing.
	logger := zerolog.Nop()
	fileObserver := &FileObserver{
		filePath:    tmp, // directory, not a file
		description: "file observer",
		logger:      &logger,
	}

	event := mustNewAuditEvent(t, "127.0.0.1", "m1")

	err := fileObserver.Update(context.Background(), event)
	if err == nil {
		t.Fatalf("expected error when appending to directory path, got nil")
	}
}

func TestFileObserver_Name(t *testing.T) {
	// ensure Name returns the description
	logger := zerolog.Nop()
	fobs := &FileObserver{
		filePath:    "/tmp/some",
		description: "my-file-obs",
		logger:      &logger,
	}

	if fobs.Name() != "my-file-obs" {
		t.Fatalf("unexpected Name: %s", fobs.Name())
	}
}
