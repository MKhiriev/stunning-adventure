package store

import (
	"context"
	"fmt"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type Database interface {
	Migrate(context.Context) error
}

type DB struct {
	*pgx.Conn
	logger *zerolog.Logger
}

func NewConnectPostgres(cfg *config.ServerConfig, log *zerolog.Logger) (*DB, error) {
	ctx := context.Background()
	// establish connection
	conn, err := pgx.Connect(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("error occured during database connection: %w", err)
	}
	log.Info().Msg("connected to database successfully")
	// construct a DB struct
	db := &DB{Conn: conn, logger: log}

	if err := db.Migrate(ctx); err != nil {
		// if there is no `metrics` table then there is no need to use db
		return nil, err
	}

	return db, nil
}

const (
	addCounter = `INSERT INTO metrics (name, type, delta) VALUES ($1, $2, $3)`
	getMetric  = `SELECT * FROM metrics m WHERE m.name = 'metr' AND m.type = 'gauge' ORDER BY m.id DESC LIMIT 1;`
)

func (db *DB) AddCounter(ctx context.Context, metrics models.Metrics) (models.Metrics, error) {
	var metric models.Metrics
	_, err := db.Exec(ctx, addCounter)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.AddCounter").Msg("error occurred during inserting metrics value")
		return models.Metrics{}, fmt.Errorf("error occurred during inserting metrics value %w", err)
	}

	//if err :=

	return metric, nil
}

func (db *DB) UpdateGauge(ctx context.Context, metrics models.Metrics) (models.Metrics, error) {
	//TODO implement me
	panic("implement me")
}

func (db *DB) GetMetricByNameAndType(ctx context.Context, metricName string, metricType string) (models.Metrics, bool) {
	//TODO implement me
	panic("implement me")
}

func (db *DB) GetAllMetrics(ctx context.Context) []models.Metrics {

	//TODO implement me
	panic("implement me")
}

func (db *DB) Migrate(ctx context.Context) error {
	query := `CREATE TABLE IF NOT EXISTS metrics (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    delta INT,
    value DOUBLE PRECISION
);`
	_, err := db.Exec(ctx, query)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.Migrate").Msg("error while creating `metrics` table")
		return fmt.Errorf("error while creating metrics table: %w", err)
	}

	return nil
}
