package storage

import (
	"fmt"
	"strconv"

	"github.com/sejo412/ya-metrics/internal/models"
)

type MemoryStorage struct {
	metrics map[string]models.Metric
}

func NewMemoryStorage() *MemoryStorage {
	metrics := make(map[string]models.Metric)
	return &MemoryStorage{metrics: metrics}
}

func (s *MemoryStorage) AddOrUpdate(metric models.Metric) error {
	if metric.Kind == models.MetricKindCounter {
		if m, ok := s.metrics[metric.Name]; ok {
			currentInt, err := strconv.Atoi(m.Value)
			if err != nil {
				return fmt.Errorf("could not convert saved metric '%s' to int", metric.Name)
			}
			newInt, err := strconv.Atoi(metric.Value)
			if err != nil {
				return fmt.Errorf("could not convert new metric '%s' to int", metric.Name)
			}
			currentInt += newInt
			metric.Value = strconv.Itoa(currentInt)
		}
	}
	s.metrics[metric.Name] = metric
	return nil
}

func (s *MemoryStorage) Get(kind, name string) (models.Metric, error) {
	// kind not used in this implementation, because name is "primary key" for MemoryStorage
	if metric, ok := s.metrics[name]; ok {
		return metric, nil
	}
	return models.Metric{}, models.ErrHTTPNotFound
}

func (s *MemoryStorage) GetAll() []models.Metric {
	metrics := make([]models.Metric, 0, len(s.metrics))
	for _, metric := range s.metrics {
		metrics = append(metrics, metric)
	}
	return metrics
}
