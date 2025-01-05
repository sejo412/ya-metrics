package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sejo412/ya-metrics/internal/models"
)

const (
	TblGauges   = "metric_gauges"
	TblCounters = "metric_counters"
	TblMapping  = "metric_mapping"
)

type PostgresStorage struct {
	Client *sql.DB
}

func NewPostgresStorage() *PostgresStorage {
	return &PostgresStorage{}
}

func (p *PostgresStorage) AddOrUpdate(ctx context.Context, metric models.Metric) error {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()

	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}

	query, args, err := postgresUpsertQueryByMetric(metric)
	if err != nil {
		return fmt.Errorf("could not construct query: %w", err)
	}

	tx, err := p.Client.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("failed to insert/update metric: %w", err)
	}
	return nil
}

func (p *PostgresStorage) MassAddOrUpdate(ctx context.Context, metrics []models.Metric) error {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()

	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}

	tx, err := p.Client.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	for _, metric := range metrics {
		query, args, err := postgresUpsertQueryByMetric(metric)
		if err != nil {
			return fmt.Errorf("could not construct query: %w", err)
		}
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to insert/update metric: %w", err)
		}
	}
	return nil
}

func (p *PostgresStorage) Get(ctx context.Context, kind, name string) (models.Metric, error) {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()
	var query string
	switch kind {
	// shut up linter
	case models.MetricKindGauge:
		query = fmt.Sprintf(`
		SELECT m.name, $1 AS type, t.value::TEXT AS value
		FROM %s m
		JOIN %s t ON m.id = t.metric_id
		WHERE m.name = $2;`,
			TblMapping, TblGauges)
	case models.MetricKindCounter:
		query = fmt.Sprintf(`
		SELECT m.name, $1 AS type, t.value::TEXT AS value
		FROM %s m
		JOIN %s t ON m.id = t.metric_id
		WHERE m.name = $2;`,
			TblMapping, TblCounters)
	default:
		return models.Metric{}, fmt.Errorf("invalid metric kind")
	}

	var value string
	if err := p.Client.QueryRowContext(ctx, query, kind, name).Scan(&name, &kind, &value); err != nil {
		if err == sql.ErrNoRows {
			return models.Metric{}, errors.New("metric not found")
		}
		return models.Metric{}, fmt.Errorf("failed to query: %w", err)
	}
	return models.Metric{
		Kind:  kind,
		Name:  name,
		Value: value,
	}, nil
}

func (p *PostgresStorage) GetAll(ctx context.Context) ([]models.Metric, error) {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()
	metrics := make([]models.Metric, 0)
	query := fmt.Sprintf(`
		SELECT m.name, $1 AS type, g.value::TEXT AS value
		FROM %s m
		JOIN %s g ON m.id = g.metric_id
		UNION ALL
		SELECT m.name, $2 AS type, c.value::TEXT AS value
		FROM %s m
		JOIN %s c ON m.id = c.metric_id;`,
		TblMapping, TblGauges, TblMapping, TblCounters)
	rows, err := p.Client.QueryContext(ctx, query, models.MetricKindGauge, models.MetricKindCounter)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()
	for rows.Next() {
		var metric models.Metric
		if err = rows.Scan(&metric.Name, &metric.Kind, &metric.Value); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}
		metrics = append(metrics, metric)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate: %w", err)
	}
	return metrics, nil
}

func (p *PostgresStorage) Flush(dst io.Writer) error {
	// not implemented yet
	return nil
}

func (p *PostgresStorage) Load(src io.Reader) error {
	// not implemented yet
	return nil
}

// Open tries open connections to database with retry
func (p *PostgresStorage) Open(ctx context.Context, opts Options) error {
	ps := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.Username, opts.Password, opts.Database, opts.SSLMode)
	var lastErr error
	for attempt := 0; attempt < models.RetryMaxRetries; attempt++ {
		db, err := sql.Open("pgx", ps)
		if err == nil {
			p.Client = db
			return nil
		}
		lastErr = err
		delay := models.RetryInitDelay + time.Duration(attempt)*models.RetryDeltaDelay
		time.Sleep(delay)
		continue
	}
	return fmt.Errorf("failed to open postgres connection: %w", lastErr)
}

func (p *PostgresStorage) Close() {
	_ = p.Client.Close()
}

// Ping tries ping with retry
// we want ping before init scheme, upsert only and using /ping location
// user doesn't want waiting 1s + 3s + 5s before getting error
// we don't interesting what happens with database
func (p *PostgresStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()
	var lastErr error
	for attempt := 0; attempt < models.RetryMaxRetries; attempt++ {
		err := p.Client.PingContext(ctx)
		if err == nil {
			return nil
		}
		lastErr = err
		delay := models.RetryInitDelay + time.Duration(attempt)*models.RetryDeltaDelay
		time.Sleep(delay)
		continue
	}
	return fmt.Errorf("error: All attempts failed [%d], last error is: %w", models.RetryMaxRetries, lastErr)
}

func (p *PostgresStorage) Init(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()
	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}
	for _, row := range postgresScheme() {
		if _, err := p.Client.ExecContext(ctx, row); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}
	return nil
}

func postgresScheme() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS ` + TblMapping + ` (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name VARCHAR(60) NOT NULL UNIQUE
		);`,

		`CREATE TABLE IF NOT EXISTS ` + TblGauges + ` (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			metric_id INTEGER NOT NULL UNIQUE,
			value DOUBLE PRECISION NOT NULL,
			FOREIGN KEY (metric_id) REFERENCES ` + TblMapping + ` (id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS ` + TblCounters + ` (
			id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			metric_id INTEGER NOT NULL UNIQUE,
			value BIGINT NOT NULL,
			FOREIGN KEY (metric_id) REFERENCES ` + TblMapping + ` (id) ON DELETE CASCADE
		);`,
	}
}

func postgresUpsertQueryWithSetValue(targetTable, setValue string) string {
	return fmt.Sprintf(`
		WITH metric AS (
			INSERT INTO %s (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
			RETURNING id
		)
		INSERT INTO %s (metric_id, value)
		VALUES (
			COALESCE((SELECT id FROM metric), (SELECT id FROM %s WHERE name = $1)), $2
		)
		ON CONFLICT (metric_id) DO UPDATE
		SET value = %s ;`, TblMapping, targetTable, TblMapping, setValue)
}

func postgresUpsertQueryByMetric(metric models.Metric) (query string, args []interface{}, err error) {
	switch metric.Kind {
	case models.MetricKindCounter:
		newCounter, err := strconv.Atoi(metric.Value)
		if err != nil {
			return "", nil, fmt.Errorf("invalid counter value: %w", err)
		}
		query = postgresUpsertQueryWithSetValue(TblCounters, TblCounters+".value + EXCLUDED.value")
		args = append(args, metric.Name, newCounter)
	case models.MetricKindGauge:
		query = postgresUpsertQueryWithSetValue(TblGauges, "EXCLUDED.value")
		args = append(args, metric.Name, metric.Value)
	default:
		return "", nil, fmt.Errorf("invalid metric kind")
	}
	return query, args, nil
}
