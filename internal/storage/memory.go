package storage

import (
	"fmt"
	"github.com/sejo412/ya-metrics/internal/config"
	"strconv"
)

type MemoryStorage struct {
	Metrics map[string]Metric
}

func NewMemoryStorage() *MemoryStorage {
	metrics := make(map[string]Metric)
	return &MemoryStorage{metrics}
}

func (s *MemoryStorage) AddOrUpdate(metric Metric) error {
	if metric.Kind == config.MetricNameCounter {
		if m, ok := s.Metrics[metric.Name]; ok {
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
	s.Metrics[metric.Name] = metric
	return nil
}

func (s *MemoryStorage) Get(name string) (Metric, error) {
	if metric, ok := s.Metrics[name]; ok {
		return metric, nil
	}
	return Metric{}, fmt.Errorf("metric not found")
}

func (s *MemoryStorage) GetAll() []Metric {
	metrics := make([]Metric, 0, len(s.Metrics))
	for _, metric := range s.Metrics {
		metrics = append(metrics, metric)
	}
	return metrics
}
