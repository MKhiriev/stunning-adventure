package store

import (
	"context"
	"fmt"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type DB struct {
	*pgx.Conn
	logger *zerolog.Logger
}

func NewConnectPostgres(cfg *config.ServerConfig, log *zerolog.Logger) (*DB, error) {
	conn, err := pgx.Connect(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("error occured during database connection: %w", err)
	}

	log.Info().Msg("connected to database successfully")
	return &DB{
		Conn:   conn,
		logger: log,
	}, nil
}
