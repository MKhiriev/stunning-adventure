package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		filePath  string
		expectErr bool
	}{
		{
			name:      "empty path",
			filePath:  "",
			expectErr: true,
		},
		{
			name:      "create file in existing directory",
			filePath:  filepath.Join(tmpDir, "file.txt"),
			expectErr: false,
		},
		{
			name:      "create file with nested directories",
			filePath:  filepath.Join(tmpDir, "a", "b", "c", "file.txt"),
			expectErr: false,
		},
		{
			name:      "file already exists",
			filePath:  filepath.Join(tmpDir, "existing.txt"),
			expectErr: false,
		},
	}

	// pre-create file for "file already exists"
	existing := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existing, []byte("data"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateFile(tt.filePath)

			if tt.expectErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.expectErr {
				if _, statErr := os.Stat(tt.filePath); statErr != nil {
					t.Fatalf("file was not created: %v", statErr)
				}
			}
		})
	}
}

func TestCreateFile_DirectoryCreationError(t *testing.T) {
	tmpDir := t.TempDir()

	// create file where directory should be
	blockingFile := filepath.Join(tmpDir, "notadir")
	if err := os.WriteFile(blockingFile, []byte("x"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	target := filepath.Join(blockingFile, "file.txt")
	err := CreateFile(target)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "error creating directory") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestAppendToFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "append.txt")

	err := AppendToFile(filePath, []byte("first"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = AppendToFile(filePath, []byte("second"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "first\nsecond\n"
	if string(content) != expected {
		t.Fatalf("unexpected content:\nwant: %q\ngot:  %q", expected, string(content))
	}
}

func TestAppendToFile_OpenError(t *testing.T) {
	tmpDir := t.TempDir()

	// directory instead of file → OpenFile error
	err := AppendToFile(tmpDir, []byte("data"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
