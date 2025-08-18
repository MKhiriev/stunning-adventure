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
	insertGaugeQuery = `INSERT INTO metrics (id, type, value)  
VALUES ($1, $2, $3)  
ON CONFLICT (id, type) DO  
UPDATE SET value = $3  
RETURNING *`
	insertCounterQuery = `INSERT INTO metrics (id, type, delta)  
VALUES ($1, $2, $3)  
ON CONFLICT (id, type) DO  
UPDATE SET delta = EXCLUDED.delta + $3  
RETURNING *`
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
		log.Err(err).Msg("error connecting database (ping)")
		return nil, err
	}
	log.Info().Msg("connected to database successfully")
	// construct a DB struct
	db := &DB{DB: conn, logger: log}

	if err := db.Migrate(ctx); err != nil {
		// if there is no `metrics` table then there is no need to use db
		return nil, err
	}

	return db, nil
}

// TODO test function
func (db *DB) Save(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	// get query for UPSERT of metrics value
	query, err := db.insertQuery(&metric)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.Save").Msg("error during preparing query for UPSERT operation")
		return models.Metrics{}, err
	}

	// prepare context
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.Save").Str("query", query).Msg("error during preparing context")
		return models.Metrics{}, err
	}
	defer stmt.Close()

	var row *sql.Row
	// execute UPSERT query
	// query also returns updated row values
	switch metric.MType {
	case models.Gauge:
		row = stmt.QueryRowContext(ctx, metric.ID, metric.MType, metric.Value)
	case models.Counter:
		row = stmt.QueryRowContext(ctx, metric.ID, metric.MType, metric.Delta)
	}

	if row == nil {
		db.logger.Err(err).Str("func", "*DB.Save").Msg("error: row is nil")
		return models.Metrics{}, nil
	}

	// scan resulting values from query
	if err = row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
		db.logger.Err(err).Str("func", "*DB.Save").Msg("error during scanning row")
		return models.Metrics{}, err
	}

	return metric, nil
}

// TODO test function
func (db *DB) SaveAll(ctx context.Context, metrics []models.Metrics) error {
	// begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.SaveAll").Msg("error during opening transaction")
		return fmt.Errorf("error during opening transaction: %w", err)
	}
	defer tx.Rollback()

	gaugeStmt, err := tx.PrepareContext(ctx, insertGaugeQuery)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.SaveAll").Msg("error creating transaction statement for Gauge UPSERT")
		return fmt.Errorf("error creating transaction statement for Gauge UPSERT: %w", err)
	}
	counterStmt, err := tx.PrepareContext(ctx, insertCounterQuery)
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.SaveAll").Msg("error creating transaction statement for Counter UPSERT")
		return fmt.Errorf("error creating transaction statement for Counter UPSERT: %w", err)
	}

	// for each metric
	for _, metric := range metrics {
		// save metric
		var result sql.Result
		var statementExecutionError error
		switch metric.MType {
		case models.Gauge:
			db.logger.Info().Str("func", "*DB.SaveAll").Any("gauge-metric", metric).Msg("trying to save gauge metric")
			result, statementExecutionError = gaugeStmt.ExecContext(ctx, metric.ID, metric.MType, metric.Value)
			if statementExecutionError != nil {
				db.logger.Err(err).Str("func", "*DB.SaveAll").Any("gauge-metric", metric).Msg("error executing prepared UPSERT query for Gauge metric")
				return err
			}
		case models.Counter:
			db.logger.Info().Str("func", "*DB.SaveAll").Any("counter-metric", metric).Msg("trying to save counter metric")
			result, statementExecutionError = counterStmt.ExecContext(ctx, metric.ID, metric.MType, metric.Delta)
			if statementExecutionError != nil {
				db.logger.Err(err).Str("func", "*DB.SaveAll").Any("counter-metric", metric).Msg("error executing prepared UPSERT query for Counter metric")
				return err
			}
		default:
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
	// prepare context
	stmt, err := db.PrepareContext(ctx, "SELECT * FROM metrics WHERE id=$1 AND type=$2")
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.Get").Msg("error during preparing context")
		return models.Metrics{}, false
	}
	defer stmt.Close()

	db.logger.Info().Str("func", "*DB.Get").Any("metric to find", metric).Msg("trying to find metric")
	// query row with given name and type
	row := stmt.QueryRowContext(ctx, metric.ID, metric.MType)
	// scan resulting row
	err = row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
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

// TODO test function
func (db *DB) GetAll(ctx context.Context) ([]models.Metrics, error) {
	stmt, err := db.PrepareContext(ctx, "SELECT * FROM metrics")
	if err != nil {
		db.logger.Err(err).Str("func", "*DB.GetAll").Msg("error during preparing context")
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
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
    delta integer,    
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

func (db *DB) insertQuery(metric *models.Metrics) (string, error) {
	switch metric.MType {
	case models.Gauge:
		return insertGaugeQuery, nil
	case models.Counter:
		return insertCounterQuery, nil
	default:
		return "", errors.New("no query for given metric type")
	}
}
