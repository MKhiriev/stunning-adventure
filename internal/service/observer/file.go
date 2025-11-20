package observer

import (
	"context"
	"os"

	"github.com/MKhiriev/stunning-adventure/models"
)

type FileObserver struct {
	file        *os.File
	description string
}

// TODO
func NewFileObserver(filePath string) AuditObserver {
	// todo add check if file exists, if not - create one
	return &FileObserver{file: nil, description: "file observer"}
}

// TODO
func (f *FileObserver) Update(ctx context.Context, event models.AuditEvent) error {
	//TODO implement me
	panic("implement me")
}

// DONE!
func (f *FileObserver) Name() string {
	return f.description
}
