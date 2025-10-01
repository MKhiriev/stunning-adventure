package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
	errorClassificator ErrorClassificator
	logger             *zerolog.Logger
	retryIntervals     map[int]time.Duration
	maxAttempts        int
}

func NewConnectPostgres(ctx context.Context, cfg *config.ServerConfig, log *zerolog.Logger) (*DB, error) {
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
	db := &DB{
		DB:                 conn,
		logger:             log,
		errorClassificator: NewPostgresErrorClassifier(),
		// add retry mechanism
		maxAttempts: 4,
		retryIntervals: map[int]time.Duration{
			1: 1 * time.Second,
			2: 3 * time.Second,
			3: 5 * time.Second,
		}}

	if err := db.Migrate(ctx); err != nil {
		log.Err(err).Str("func", "NewConnectPostgres").Msg("failed migration")
		// if there is no `metrics` table then there is no need to use db
		return nil, err
	}

	return db, nil
}

func (db *DB) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	var result models.Metrics
	err := db.withRetry(ctx, "*DB.Save", func() error {
		var saveErr error
		result, saveErr = db.saveMetric(ctx, metric)
		return saveErr
	})
	return result, err
}

func (db *DB) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	err := db.withRetry(ctx, "*DB.SaveAll", func() error {
		saveAllErr := db.saveAllMetrics(ctx, metrics)
		return saveAllErr
	})
	return err
}

func (db *DB) Get(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	var foundMetric models.Metrics
	err := db.withRetry(ctx, "*DB.Get", func() error {
		var getErr error
		foundMetric, getErr = db.getMetric(ctx, metric)
		return getErr
	})
	if err != nil {
		return models.Metrics{}, err
	}

	return foundMetric, nil
}

func (db *DB) GetAll(ctx context.Context) ([]models.Metrics, error) {
	var result []models.Metrics
	err := db.withRetry(ctx, "*DB.GetAll", func() error {
		var getAllErr error
		result, getAllErr = db.getAllMetrics(ctx)
		return getAllErr
	})
	return result, err
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

func (db *DB) withRetry(ctx context.Context, operation string, fn func() error) error {
	var err error
	for attempt := 1; attempt <= db.maxAttempts; attempt++ {
		db.logger.Info().Str("func", operation).Int("attempt", attempt).Msg("trying operation")
		err = fn()

		if !db.checkIfRetryable(err) {
			if err != nil {
				db.logger.Err(err).Str("func", operation).Int("attempt", attempt).Msg("operation failed")
			}
			return err
		}

		time.Sleep(db.retryIntervals[attempt])
	}
	return err
}

func (db *DB) saveMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	if metric.MType == models.Gauge || metric.MType == models.Counter {
		db.logger.Info().Str("func", "*DB.saveMetric").Any("metric", metric).Msg("trying to save metric")
		// save metric in db
		row := db.QueryRowContext(ctx, insertMetricsQuery, metric.ID, metric.MType, metric.Delta, metric.Value)
		if err := row.Err(); err != nil {
			db.logger.Error().Err(err).Str("func", "*DB.saveMetric").Msg("error: row is nil")
			return models.Metrics{}, err
		}

		// scan saved metric from db
		if err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			db.logger.Error().Err(err).Str("func", "*DB.saveMetric").Msg("error: scanning error")
			return models.Metrics{}, err
		}
	} else {
		db.logger.Error().Str("func", "*DB.saveMetric").Any("metric", metric).Msg("unsupported metric type was passed")
		return models.Metrics{}, errors.New("unsupported metric type was passed")
	}

	// return saved in db metric
	return metric, nil
}

func (db *DB) saveAllMetrics(ctx context.Context, metrics []models.Metrics) error {
	// begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.saveAllMetrics").Msg("error during opening transaction")
		return fmt.Errorf("error during opening transaction: %w", err)
	}
	defer tx.Rollback()

	// prepare context
	stmt, err := tx.PrepareContext(ctx, insertMetricsQuery)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.saveAllMetrics").Msg("error during preparing context")
		return err
	}
	defer stmt.Close()

	// for each metric
	for idx, metric := range metrics {
		// save metric
		var result sql.Result
		var statementExecutionError error
		if metric.MType == models.Gauge || metric.MType == models.Counter {
			db.logger.Info().Str("func", "*DB.saveAllMetrics").Any("metric", metric).Int("iteration", idx).Msg("trying to save metric")
			result, statementExecutionError = stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta, metric.Value)
			if statementExecutionError != nil {
				db.logger.Err(statementExecutionError).Str("func", "*DB.saveAllMetrics").Any("metric", metric).Int("iteration", idx).Msg("error executing prepared UPSERT query for saving metric")
				return statementExecutionError
			}
		} else {
			db.logger.Error().Str("func", "*DB.saveAllMetrics").Any("metric", metric).Int("iteration", idx).Msg("unsupported metric type was passed")
			return errors.New("unsupported metric type was passed")
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			db.logger.Err(err).Str("func", "*DB.saveAllMetrics").Msg("metric was not updated")
			return err
		}
	}
	// commit transaction if all metrics are successfully updated

	return tx.Commit()
}

func (db *DB) getMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	db.logger.Info().Str("func", "*DB.getMetric").Any("metric to find", metric).Msg("trying to find metric")
	// query row with given name and type
	row := db.QueryRowContext(ctx, getMetric, metric.ID, metric.MType)
	// scan resulting row
	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	// check for error type
	switch {
	case errors.Is(err, sql.ErrNoRows):
		db.logger.Err(err).Str("func", "*DB.getMetric").Msg("no rows were found")
		return models.Metrics{}, ErrNotFound
	case err != nil:
		db.logger.Err(err).Str("func", "*DB.getMetric").Msg("error occurred during scanning")
		return models.Metrics{}, err
	default:
		db.logger.Info().Str("func", "*DB.getMetric").Any("found metric", metric).Msg("metric IS found")
		return metric, nil
	}
}

func (db *DB) getAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	rows, err := db.QueryContext(ctx, getAllMetrics)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.getAllMetrics").Msg("error during query execution")
		return nil, err
	}

	var all []models.Metrics
	for rows.Next() {
		var metric models.Metrics
		err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
		if err != nil {
			db.logger.Err(err).Str("func", "*DB.getAllMetrics").Msg("error during getting values from row")
			return nil, err
		}

		all = append(all, metric)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		db.logger.Err(rowsErr).Str("func", "*DB.getAllMetrics").Msg("error during rows scanning")
		return nil, rowsErr
	}

	return all, nil
}

func (db *DB) checkIfRetryable(err error) bool {
	db.logger.Info().Str("func", "*DB.checkIfRetryable").Msg("checking if given PostgreSQL error is retryable")
	if db.errorClassificator.Classify(err) == NonRetryable {
		db.logger.Info().Str("func", "*DB.checkIfRetryable").Msg("given PostgreSQL error is NOT retryable")
		return false
	}

	db.logger.Info().Str("func", "*DB.checkIfRetryable").Msg("given PostgreSQL error is retryable")
	return true
}
