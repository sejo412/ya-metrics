package storage

import "github.com/sejo412/ya-metrics/internal/domain"

type Storage interface {
	AddOrUpdate(domain.Metric) error
	Get(string) (domain.Metric, error)
	GetAll() []domain.Metric
}
