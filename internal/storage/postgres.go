package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sejo412/ya-metrics/internal/models"
)

type PostgresStorage struct {
	Client *sql.DB
}

func (p *PostgresStorage) AddOrUpdate(metric models.Metric) error {
	// TODO implement me
	panic("implement me")
}

func (p *PostgresStorage) Get(kind string, name string) (models.Metric, error) {
	// TODO implement me
	panic("implement me")
}

func (p *PostgresStorage) GetAll() []models.Metric {
	// TODO implement me
	panic("implement me")
}

func (p *PostgresStorage) Flush(dst io.Writer) error {
	// TODO implement me
	panic("implement me")
}

func (p *PostgresStorage) Load(src io.Reader) error {
	// TODO implement me
	panic("implement me")
}

func NewPostgresStorage() *PostgresStorage {
	return &PostgresStorage{}
}

func (p *PostgresStorage) Open(opts Options) error {
	ps := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		opts.Host, opts.Port, opts.Username, opts.Password, opts.Database, opts.SSLMode)
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return fmt.Errorf("failed to open postgres connection: %w", err)
	}
	p.Client = db
	return nil
}

func (p *PostgresStorage) Close() {
	_ = p.Client.Close()
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, ctxTimeout)
	defer cancel()
	return p.Client.PingContext(ctx)
}
