package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MKhiriev/stunning-adventure/internal/config"
	"github.com/MKhiriev/stunning-adventure/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

const (
	insertMetricsQuery = `INSERT INTO metrics (id, type, delta, value) 
VALUES ($1, $2, $3, $4)
ON CONFLICT (id, type) DO 
UPDATE SET 
           value = EXCLUDED.value,
           delta = metrics.delta + EXCLUDED.delta
RETURNING *;`
	getMetric     = `SELECT * FROM metrics WHERE id=$1 AND type=$2;`
	getAllMetrics = `SELECT * FROM metrics;`
)

type DB struct {
	*sql.DB
	logger *zerolog.Logger
}

func NewConnectPostgres(cfg *config.ServerConfig, log *zerolog.Logger) (*DB, error) {
	ctx := context.Background()
	// establish connection
	conn, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Err(err).Str("func", "NewConnectPostgres").Msg("error occured during database connection")
		return nil, fmt.Errorf("error occured during database connection: %w", err)
	}
	err = conn.PingContext(ctx)
	if err != nil {
		log.Err(err).Str("func", "NewConnectPostgres").Msg("error connecting database (ping)")
		return nil, err
	}
	log.Info().Str("func", "NewConnectPostgres").Msg("connected to database successfully")
	// construct a DB struct
	db := &DB{DB: conn, logger: log}

	if err := db.Migrate(ctx); err != nil {
		log.Err(err).Str("func", "NewConnectPostgres").Msg("failed migration")
		// if there is no `metrics` table then there is no need to use db
		return nil, err
	}

	return db, nil
}

func (db *DB) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	var row *sql.Row
	// execute UPSERT query
	// query also returns updated row values
	switch metric.MType {
	case models.Gauge:
		row = db.QueryRowContext(ctx, metric.ID, metric.MType, metric.Value)
	case models.Counter:
		row = db.QueryRowContext(ctx, metric.ID, metric.MType, metric.Delta)
	}

	if metric.MType == models.Gauge || metric.MType == models.Counter {
		db.logger.Info().Str("func", "*DB.Save").Any("metric", metric).Msg("trying to save metric")
		row = db.QueryRowContext(ctx, insertMetricsQuery, metric.ID, metric.MType, metric.Delta, metric.Value)

		if row == nil {
			db.logger.Error().Str("func", "*DB.Save").Msg("error: row is nil")
			return models.Metrics{}, nil
		}
	} else {
		db.logger.Error().Str("func", "*DB.SaveAll").Any("metric", metric).Msg("unsupported metric type was passed")
		return models.Metrics{}, errors.New("unsupported metric type was passed")
	}

	// scan resulting values from query
	if err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
		db.logger.Err(err).Str("func", "*DB.Save").Msg("error during scanning row")
		return models.Metrics{}, err
	}

	return metric, nil
}

func (db *DB) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	// begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.SaveAll").Msg("error during opening transaction")
		return fmt.Errorf("error during opening transaction: %w", err)
	}
	defer tx.Rollback()

	// prepare context
	stmt, err := tx.PrepareContext(ctx, insertMetricsQuery)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.SaveAll").Msg("error during preparing context")
		return err
	}
	defer stmt.Close()

	// for each metric
	for idx, metric := range metrics {
		// save metric
		var result sql.Result
		var statementExecutionError error
		if metric.MType == models.Gauge || metric.MType == models.Counter {
			db.logger.Info().Str("func", "*DB.SaveAll").Any("metric", metric).Int("iteration", idx).Msg("trying to save metric")
			result, statementExecutionError = stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value)
			if statementExecutionError != nil {
				db.logger.Err(statementExecutionError).Str("func", "*DB.SaveAll").Any("metric", metric).Int("iteration", idx).Msg("error executing prepared UPSERT query for saving metric")
				return statementExecutionError
			}
		} else {
			db.logger.Error().Str("func", "*DB.SaveAll").Any("metric", metric).Int("iteration", idx).Msg("unsupported metric type was passed")
			return errors.New("unsupported metric type was passed")
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			db.logger.Err(err).Str("func", "*DB.SaveAll").Msg("metric was not updated")
			return err
		}
	}

	// commit transaction if all metrics are successfully updated

	return tx.Commit()
}

func (db *DB) Get(ctx context.Context, metric models.Metrics) (models.Metrics, bool) {
	db.logger.Info().Str("func", "*DB.Get").Any("metric to find", metric).Msg("trying to find metric")
	// query row with given name and type
	row := db.QueryRowContext(ctx, getMetric, metric.ID, metric.MType)
	// scan resulting row
	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	// check for error type
	switch {
	case errors.Is(err, sql.ErrNoRows):
		db.logger.Err(err).Str("func", "*DB.Get").Msg("no rows were found")
		return models.Metrics{}, false
	case err != nil:
		db.logger.Err(err).Str("func", "*DB.Get").Msg("error occurred during scanning")
		return models.Metrics{}, false
	default:
		return metric, true
	}
}

func (db *DB) GetAll(ctx context.Context) ([]models.Metrics, error) {
	rows, err := db.QueryContext(ctx, getAllMetrics)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.GetAll").Msg("error during query execution")
		return nil, err
	}

	var all []models.Metrics
	for rows.Next() {
		var metric models.Metrics
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			db.logger.Err(err).Str("func", "*DB.GetAll").Msg("error during getting values from row")
			return nil, err
		}

		all = append(all, metric)
	}

	if rows.Err() != nil {
		db.logger.Err(err).Str("func", "*DB.GetAll").Msg("error during rows scanning")
		return nil, err
	}

	return all, nil
}

func (db *DB) Migrate(ctx context.Context) error {
	query := `  
create table if not exists metrics  
(  
    id  text not null,   
    type  text not null,    
    delta bigint default null,    
    value double precision,    
    primary key (id, type)
);`
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.Migrate").Msg("error while creating `metrics` table")
		return fmt.Errorf("error while creating metrics table: %w", err)
	}

	return nil
}
