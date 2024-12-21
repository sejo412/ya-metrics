package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
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

func (s *MemoryStorage) Flush(file string) error {
	f, err := os.Create(file)
	defer func() {
		_ = f.Close()
	}()
	if err != nil {
		return fmt.Errorf("error create file %s: %w", file, err)
	}
	metrics := s.GetAll()
	for _, metric := range metrics {
		m, err := models.ConvertMetricToV2(&metric)
		if err != nil {
			return err
		}
		err = json.NewEncoder(f).Encode(m)
		if err != nil {
			return fmt.Errorf("error encode metric %s: %w", metric, err)
		}
	}
	return nil
}

func (s *MemoryStorage) Load(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("error open file %s: %w", file, err)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m := models.MetricV2{}
		err = json.Unmarshal(scanner.Bytes(), &m)
		if err != nil {
			return fmt.Errorf("error unmarshal metric %s: %w", scanner.Text(), err)
		}
		res, err := models.ConvertV2ToMetric(&m)
		if err != nil {
			return err
		}
		if err = s.AddOrUpdate(*res); err != nil {
			return fmt.Errorf("error add or update metric %s: %w", res.Name, err)
		}
	}
	return nil
}
