package storage

import (
	"fmt"
	"github.com/sejo412/ya-metrics/internal/domain"
	"strconv"
)

type MemoryStorage struct {
	metrics map[string]domain.Metric
}

func NewMemoryStorage() *MemoryStorage {
	metrics := make(map[string]domain.Metric)
	return &MemoryStorage{metrics}
}

func (s *MemoryStorage) AddOrUpdate(metric domain.Metric) error {
	if metric.Kind == domain.MetricKindCounter {
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

func (s *MemoryStorage) Get(name string) (domain.Metric, error) {
	if metric, ok := s.metrics[name]; ok {
		return metric, nil
	}
	return domain.Metric{}, fmt.Errorf("metric not found")
}

func (s *MemoryStorage) GetAll() []domain.Metric {
	metrics := make([]domain.Metric, 0, len(s.metrics))
	for _, metric := range s.metrics {
		metrics = append(metrics, metric)
	}
	return metrics
}
