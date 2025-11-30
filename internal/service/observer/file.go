package observer

import (
	"context"
	"encoding/json"

	"github.com/MKhiriev/stunning-adventure/internal/utils"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/rs/zerolog"
)

type FileObserver struct {
	filePath    string
	description string
	logger      *zerolog.Logger
}

func NewFileObserver(filePath string, logger *zerolog.Logger) AuditObserver {
	if err := utils.CreateFile(filePath); err != nil {
		logger.Err(err).
			Str("func", "observer.NewFileObserver").
			Msg("error during creating observers' file")
		return nil
	}

	return &FileObserver{filePath: filePath, description: "file observer"}
}

func (f *FileObserver) Update(ctx context.Context, event models.AuditEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		f.logger.Err(err).Str("func", "observer.NewFileObserver").
			Msg("event to JSON conversion error")
		return err
	}

	return utils.AppendToFile(f.filePath, eventJSON)
}

func (f *FileObserver) Name() string {
	return f.description
}
